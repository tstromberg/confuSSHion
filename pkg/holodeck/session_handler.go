package holodeck

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/google/generative-ai-go/genai"
	"github.com/tstromberg/confuSSHion/pkg/history"
	"golang.org/x/term"
	"k8s.io/klog/v2"
)

var globalCache = map[string]*Response{}

// Handler handles SSH sessions
func (h Holodeck) Handler(s ssh.Session) error {
	sid := s.Context().SessionID()[0:12]
	ip, _, _ := strings.Cut(s.LocalAddr().String(), ":")
	// We assume IPv4, because some systems we emulate may not have IPv6 compatibility.
	if len(ip) < 7 {
		ip = "10.0.0.5"
	}

	sess := &history.SessionContext{
		SID:             sid,
		StartTime:       time.Now(),
		OS:              h.p.Name(),
		Arch:            h.p.Arch(),
		Hostname:        h.p.Hostname(),
		User:            s.User(),
		RemoteAddr:      s.RemoteAddr().String(),
		Environ:         s.Environ(),
		LoginCommand:    strings.Join(s.Command(), " "),
		RoleDescription: h.nc.RoleDescription,
		PromptHints:     h.p.Hints(),
		History:         []history.Entry{},
		LocalAddr:       s.LocalAddr().String(),
		NodeIP:          ip,
	}
	klog.Infof("session[%s@%s] from %s to %s [%s]: cmd=%v, environ=%v", sess.User, sess.SID, sess.RemoteAddr, sess.LocalAddr, ip, sess.LoginCommand, sess.Environ)

	var err error
	var resp *Response
	shellPrompt := h.p.ShellPrompt()
	time.Sleep(500 * time.Millisecond)

	if len(s.Command()) == 0 {
		resp, err = h.hallucinate(execTmpl("welcome", sess))
		if shellPrompt == "" {
			shellPrompt = resp.ShellPrompt
		}
	} else {
		cmd := strings.Join(s.Command(), " ")
		baseCmd := filepath.Base(s.Command()[0])
		resp = globalCache[cmd]
		if resp != nil {
			klog.V(1).Infof("using global cache for %q", cmd)
		} else {
			resp, err = h.hallucinate(execTmpl("login_command", sess))
			if err == nil {
				klog.Infof("Base command: %s", baseCmd)
				if globallyCacheable[baseCmd] {
					klog.Infof("globally caching %q as %s", cmd, resp.Output)
					globalCache[cmd] = resp
				}
			}
		}
	}

	if err != nil {
		time.Sleep(1 * time.Second)
		return fmt.Errorf("hallucination error: %w", err)
	}
	if resp == nil {
		return fmt.Errorf("initial response is nil?", err)
	}

	// Record the welcome message
	sess.History = append(sess.History, history.Entry{T: time.Now(), Kind: "login", In: sess.LoginCommand, Out: string(resp.Output)})

	defer func() {
		if h.historyStore == nil {
			return
		}
		sess.EndTime = time.Now()
		if err := h.historyStore.SaveSession(sess); err != nil {
			klog.Errorf("Failed to save session history: %v", err)
		}
	}()

	term := term.NewTerminal(s, shellPrompt)
	klog.Infof("login response: %s", resp.Output)
	term.Write(resp.Output)

	if len(s.Command()) > 0 {
		return nil
	}

	for {
		klog.Info("Waiting for input...")
		cmd, err := term.ReadLine()
		if err != nil {
			return fmt.Errorf("readline: %v", err)
		}

		cmd = strings.TrimSpace(cmd)
		if len(cmd) > 1024 {
			klog.Warningf("truncating long command (%d)", len(cmd))
			cmd = cmd[0:1023]
		}
		klog.Infof("Command received: [%s]", cmd)
		sess.CurrentCommand = cmd

		if cmd == "" {
			klog.Infof("<empty line>")
			continue
		}
		time.Sleep(200 * time.Millisecond)

		// Process command
		cmdBin, _, _ := strings.Cut(cmd, " ")
		baseCmd := filepath.Base(cmdBin)
		klog.Infof("Base command: %s", baseCmd)

		resp := globalCache[cmd]
		if resp != nil {
			klog.Infof("re-using globally cached result for %s", cmd)
		} else {
			klog.V(1).Infof("%q not in global cache", cmd)
		}

		if globallyCacheable[baseCmd] || locallyCacheable[baseCmd] {
			klog.V(1).Infof("%q is cacheable!", baseCmd)
			if cached := h.cache[cmd]; cached != nil {
				klog.Infof("re-using locally cached result for %s", cmd)
				resp = cached
			}
		} else {
			klog.Infof("Flushing local cache for non-cacheable command: %q", baseCmd)
			h.cache = make(map[string]*Response)
		}

		// Generate response if not cached
		if resp == nil {
			klog.V(1).Infof("uncached, history length: %d", len(sess.History))
			if len(sess.History) == 1 {
				resp, err = h.hallucinate(execTmpl("first_command", sess))
			} else {
				resp, err = h.hallucinate(execTmpl("subsequent_commands", sess))
			}

			klog.Infof("is %q cacheable globally? %v", baseCmd, globallyCacheable[baseCmd])
			if globallyCacheable[baseCmd] {
				klog.Infof("globally caching %q", cmd)
				globalCache[cmd] = resp
			} else if locallyCacheable[baseCmd] {
				klog.Infof("locally caching %q", baseCmd)
				h.cache[cmd] = resp
			}
		}

		if resp == nil {
			klog.Error("response is nil? oops!")
			continue
		}

		term.Write(resp.Output)
		sess.History = append(sess.History, history.Entry{Kind: "cmd", T: time.Now(), In: cmd, Out: string(resp.Output)})

		if baseCmd == "cd" {
			sess.CurrentWorkingDirectory = ""
			cdArgs := strings.Split(cmd, " ")
			klog.Infof("args=%v len=%d", cdArgs, len(bytes.TrimSpace(resp.Output)))
			if len(cdArgs) > 1 && len(bytes.TrimSpace(resp.Output)) <= 1 && !strings.Contains(cmd, "$") {
				klog.Infof("updating current working directory to %s", cdArgs[1])
				sess.CurrentWorkingDirectory = cdArgs[1]
			}
		}

		if baseCmd == "exit" || baseCmd == "logout" || baseCmd == "reboot" || baseCmd == "LOGOUT" || baseCmd == "shutdown" {
			time.Sleep(time.Second)
			break
		}

		if cmd == "kill -9 -1" {
			time.Sleep(time.Second)
			break

		}

		if baseCmd == "cd" && h.p.ShellPrompt() == "" && strings.TrimSpace(resp.ShellPrompt) != "" {
			klog.Infof("Command=%q - setting prompt to: %q", cmd, resp.ShellPrompt)
			term.SetPrompt(resp.ShellPrompt)
		}
	}

	return nil
}

func execTmpl(name string, sess *history.SessionContext) string {
	var sb strings.Builder
	if err := templates.ExecuteTemplate(&sb, fmt.Sprintf("%s.tmpl", name), sess); err != nil {
		klog.Errorf("TEMPLATE FAIL: %q - %v", name, err)
		return ""
	}
	return sb.String()
}

// hallucinate generates a response for the given prompt using the LLM
func (h Holodeck) hallucinate(prompt string) (*Response, error) {
	er := &Response{Output: []byte("Fork failed: Resource temporarily unavailable\n"), ShellPrompt: "$"}

	if prompt == "" {
		return er, nil
	}

	klog.Infof("Hallucinating content for a %d byte prompt ...", len(prompt))
	resp, err := h.model.GenerateContent(h.ctx, genai.Text(prompt))
	if err != nil {
		return er, fmt.Errorf("model generation failed: %w", err)
	}

	// Process the response
	raw := strings.TrimSpace(fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0]))
	var output []string
	var shellPrompt string

	lines := []string{}
	for _, l := range strings.Split(raw, "\n") {
		// Skip markdown formatting (vertex bug workaround)
		if strings.HasPrefix(l, "```") || strings.HasSuffix(l, "```") {
			continue
		}
		if strings.HasPrefix(l, "`") {
			l = strings.Trim(l, "`")
		}
		if strings.HasPrefix(l, "[") {
			l = strings.Trim(l, "[]")
		}
		lines = append(lines, l)
	}

	for x, l := range lines {
		klog.Infof("%02d| %q", x, l)
		if shellPrompt == "" && x >= len(lines)-1 {
			if shellPromptRe.MatchString(l) {
				klog.Infof("Found shell prompt: %q", l)
				output = append(output, "")
				shellPrompt = l
				break
			} else {
				klog.V(1).Infof("Expected prompt not found in: %q", l)
			}
		}
		output = append(output, l)
	}

	out := strings.TrimSpace(strings.Join(output, "\n"))
	if len(out) > 0 {
		out += "\n"
	}
	// Ensure proper formatting
	if !strings.HasSuffix(shellPrompt, " ") {
		shellPrompt = shellPrompt + " "
	}

	return &Response{
		Output:      []byte(out),
		ShellPrompt: shellPrompt,
	}, nil
}
