package holodeck

import (
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

// Handler handles SSH sessions
func (h Holodeck) Handler(s ssh.Session) error {
	sid := s.Context().SessionID()

	sess := &history.SessionContext{
		SID:             sid,
		StartTime:       time.Now(),
		OS:              h.p.Name(),
		Arch:            h.p.Arch(),
		Hostname:        h.p.Hostname(),
		User:            s.User(),
		RemoteAddr:      s.RemoteAddr().String(),
		Environ:         s.Environ(),
		LoginCommand:    s.Command(),
		RoleDescription: h.nc.RoleDescription,
		PromptHints:     h.p.Hints(),
		History:         []history.Entry{},
	}
	klog.Infof("new session: %+v", sess)

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
		resp, err = h.hallucinate(execTmpl("login_command", sess))
	}

	if err != nil {
		time.Sleep(1 * time.Second)
		return fmt.Errorf("hallucination error: %w", err)
	}

	// Record the welcome message
	sess.History = append(sess.History, history.Entry{T: time.Now(), Kind: "login", In: strings.Join(s.Command(), " "), Out: string(resp.Output)})

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
	term.Write(resp.Output)

	if len(s.Command()) > 0 {
		return nil
	}

	for {
		klog.Info("Waiting for user input...")
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
		var resp *Response

		if cacheable[baseCmd] {
			klog.Infof("%q is cacheable!", baseCmd)
			if cached := h.cache[cmd]; cached != nil {
				klog.Infof("re-using cached result for %s", cmd)
				resp = cached
			}
		} else {
			klog.Infof("Flushing cache for non-cacheable command: %q", baseCmd)
			h.cache = make(map[string]*Response)
		}

		klog.Infof("post-cache resp: %v", resp)
		// Generate response if not cached
		if resp == nil {
			klog.Infof("uncached, history length: %d", len(sess.History))
			if len(sess.History) == 1 {
				resp, err = h.hallucinate(execTmpl("first_command", sess))
			} else {
				resp, err = h.hallucinate(execTmpl("subsequent_commands", sess))
			}

			if cacheable[baseCmd] {
				h.cache[cmd] = resp
			}
		}

		if resp == nil {
			klog.Error("response is nil? oops!")
			continue
		}

		term.Write(resp.Output)
		sess.History = append(sess.History, history.Entry{Kind: "cmd", T: time.Now(), In: cmd, Out: string(resp.Output)})

		if baseCmd == "exit" || baseCmd == "logout" || baseCmd == "reboot" || baseCmd == "LOGOUT" {
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

	klog.V(1).Infof("Sending prompt: %q", prompt)
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
		lines = append(lines, l)
	}

	for x, l := range lines {
		klog.Infof("Line %02d: %q", x, l)
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
