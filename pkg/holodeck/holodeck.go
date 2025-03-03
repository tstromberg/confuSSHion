package holodeck

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"

	"github.com/gliderlabs/ssh"
	"github.com/tstromberg/confuSSHion/pkg/auth"
	"github.com/tstromberg/confuSSHion/pkg/history"
	"github.com/tstromberg/confuSSHion/pkg/personality"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"k8s.io/klog/v2"
)

var shellPromptRe = regexp.MustCompile(`\$|\>|\%|\#`)

// cacheableCommand helps with ensuring that command output is consistent
var cacheableCommand = map[string]bool{
	"ls":         true,
	"id":         true,
	"w":          true,
	"last":       true,
	"ps":         true,
	"clear":      true,
	"cat":        true,
	"echo":       true,
	"less":       true,
	"man":        true,
	"whoami":     true,
	"grep":       true,
	"head":       true,
	"tail":       true,
	"diff":       true,
	"cmp":        true,
	"comm":       true,
	"sort":       true,
	"df":         true,
	"ifconfig":   true,
	"whereis":    true,
	"whatis":     true,
	"apropos":    true,
	"top":        true,
	"pwd":        true,
	"wc":         true,
	"find":       true,
	"uptime":     true,
	"du":         true,
	"awk":        true,
	"sed":        true,
	"uniq":       true,
	"htop":       true,
	"iostat":     true,
	"vmstat":     true,
	"netstat":    true,
	"ss":         true,
	"dig":        true,
	"traceroute": true,
	"nslookup":   true,
	"ip":         true,
	"route":      true,
	"hostname":   true,
	"domainname": true,
	"lscpu":      true,
	"lsblk":      true,
	"lsmod":      true,
	"dmesg":      true,
	"nc":         true,
}

// New returns a new holodeck
func New(ctx context.Context, model *genai.GenerativeModel, nc personality.NodeConfig, histStore *history.Store, a auth.Authenticator) Holodeck {
	return Holodeck{
		nc:           nc,
		model:        model,
		ctx:          ctx,
		sess:         map[string]*auth.UserInfo{},
		historyStore: histStore,
		auth:         a,
		cache:        map[string]*Response{},
	}
}

type Holodeck struct {
	// ctx shouldn't be here, but the alternative approaches suck
	ctx          context.Context
	nc           personality.NodeConfig
	model        *genai.GenerativeModel
	history      []string
	p            personality.Personality
	auth         auth.Authenticator
	historyStore *history.Store

	sess  map[string]*auth.UserInfo
	cache map[string]*Response
}

type Response struct {
	Output      []byte
	ShellPrompt string
	Logout      bool
}

func (h Holodeck) Handler(s ssh.Session) error {
	sid := s.Context().SessionID()
	userInfo := personality.UserInfo{
		RemoteUser: s.User(),
		AuthUser:   h.sess[sid],
		RemoteAddr: s.RemoteAddr().String(),
		Environ:    s.Environ(),
		Command:    s.Command(),
	}

	h.p = personality.New(h.nc, userInfo)

	// Initialize session history
	hs := &history.Session{
		SID:        sid,
		StartTime:  time.Now(),
		UserInfo:   userInfo,
		NodeConfig: h.nc,
		Log:        []history.Entry{},
	}

	resp, err := h.simulate("__login__")
	if err != nil {
		return err
	}

	shp := h.p.ShellPrompt()
	if shp == "" {
		shp = resp.ShellPrompt
	}

	// Record the welcome message
	hs.Log = append(hs.Log, history.Entry{
		Timestamp: time.Now(),
		Input:     "LOGIN",
		Output:    string(resp.Output),
	})

	time.Sleep(500 * time.Millisecond)
	term := term.NewTerminal(s, shp)
	term.Write(resp.Output)
	cmdHistory := []string{}

	defer func() {
		// Set end time and save the session history
		hs.EndTime = time.Now()
		if h.historyStore != nil {
			if err := h.historyStore.SaveSession(hs); err != nil {
				klog.Errorf("Failed to save session history: %v", err)
			} else {
				klog.Infof("Successfully saved session history for %s", sid)
			}
		}
	}()

	for {
		// Read command from SSH session
		klog.Infof("waiting for user input ...")
		cmd, err := term.ReadLine()
		if err != nil {
			klog.Errorf("readline: %v", err)
			break
		}

		klog.Infof("cmd: [%s]", strings.TrimSpace(cmd))
		time.Sleep(200 * time.Millisecond)
		if cmd == "" {
			continue
		}

		var resp *Response
		cmdBin, _, _ := strings.Cut(cmd, " ")
		baseCmd := filepath.Base(cmdBin)
		klog.Infof("base cmd: %s", baseCmd)

		// NOTE: this cache isn't yet ideally configured
		if cacheableCommand[baseCmd] {
			if h.cache[cmd] != nil {
				klog.Infof("re-using cached command: %q", cmd)
				resp = h.cache[cmd]
			}
		} else {
			klog.Infof("flushing cache for %q", baseCmd)
			h.cache = map[string]*Response{}
		}

		if resp == nil {
			prompt := fmt.Sprintf(`The first command this user is invoking in their interactive SSH session is %q
Generate the appropriate output for that command. Your response will be sent literally to the user, so do not return any markdown specific output.`, cmd)

			if len(cmdHistory) > 0 {
				prompt = fmt.Sprintf(`The user just entered the command: %q
The user has already entered %d commands, here they are in ascending time order:
"%s"

- Generate the appropriate output for that command, incorporating any state changes that previous commands may have caused.
- If the user has previously created a directory via mkdir - let them cd to it or rmdir it
- If the user has previously created a file via touch or other commands - let them remove it (the rm command should work and produce no output)
- rmdir should always succeed in deleting directories
- the rm should always succeed in deleting files, and never output "no such file or directory"
- the curl command should always appear succeed
- the wget command should always appear to succeed
- if the user tries to execute a program they have downloaded, after running chmod on it, you should fail to run it due to an architectural mismatch.
- Unless the user has changed their current working directory using the 'cd' command, assume their current working directory is their home directory.

Your response will be sent literally to the user, so do not return any markdown specific output.
`, cmd, len(cmdHistory), strings.Join(cmdHistory, "\"\n\""))
			}
			resp, err = h.simulate(prompt)
		}

		if cacheableCommand[baseCmd] && h.cache[cmd] == nil {
			klog.Infof("caching output of %q", cmd)
			h.cache[cmd] = resp
		}

		if resp == nil {
			klog.Errorf("resp is nil for %q? that isn't good!", cmd)
			continue
		}
		term.Write(resp.Output)

		// Record command and response in history
		hs.Log = append(hs.Log, history.Entry{
			Timestamp: time.Now(),
			Input:     cmd,
			Output:    string(resp.Output),
		})

		if baseCmd == "exit" || baseCmd == "logout" || baseCmd == "reboot" {
			time.Sleep(1 * time.Second)
			break
		}

		if baseCmd == "cd" && h.p.ShellPrompt() == "" && strings.TrimSpace(resp.ShellPrompt) != "" {
			klog.Infof("cmd=%q - setting prompt to: %q", cmd, resp.ShellPrompt)
			term.SetPrompt(resp.ShellPrompt)
		}
		cmdHistory = append(cmdHistory, cmd)
	}

	return nil
}

func (h Holodeck) simulate(prompt string) (*Response, error) {
	fullPrompt := ""
	if prompt == "__login__" {
		fullPrompt = h.p.WelcomePrompt()
	} else {
		fullPrompt = h.p.CommandPrompt() + "\n\n" + prompt
	}
	klog.Infof("sending prompt: %s", fullPrompt)
	resp, err := h.model.GenerateContent(h.ctx, genai.Text(fullPrompt))
	if err != nil {
		return nil, err
	}

	output := []string{}
	shellPrompt := ""
	raw := strings.TrimSpace(fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0]))

	lines := []string{}
	for _, l := range strings.Split(raw, "\n") {
		// vertex bug workaround, sometimes it returns markdown
		if strings.HasPrefix(l, "`") {
			continue
		}
		if strings.HasPrefix(l, "`") {
			log.Printf("skipping markdown line: %q", l)
			continue
		}
		lines = append(lines, l)
	}

	for x, l := range lines {
		klog.Infof("%2.2d | %q", x, l)
		if shellPrompt == "" && x >= len(lines)-1 {
			if shellPromptRe.MatchString(l) {
				klog.Infof("found shell prompt: %q", l)
				output = append(output, "")
				shellPrompt = l
				break
			} else {
				klog.Errorf("did not find expected prompt: %q", l)
			}
		}
		output = append(output, l)
	}

	out := strings.Join(output, "\n")
	if !strings.HasSuffix(shellPrompt, " ") {
		shellPrompt = shellPrompt + " "
	}

	if !strings.HasSuffix(out, "\n") {
		out = out + "\n"
	}

	return &Response{
		Output:      []byte(out),
		ShellPrompt: shellPrompt,
	}, nil
}

func (h Holodeck) PublicKeyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	klog.Infof("public key handler")
	txt := strings.TrimSpace(string(gossh.MarshalAuthorizedKey(key)))
	klog.Infof("session %s key provided: [marshal=%s]", ctx.SessionID(), txt)
	user := h.auth.ValidKey(txt)
	if user == nil {
		klog.Errorf("unable to validate %s", key)
		return false
	}

	klog.Infof("authenticated user: %s", user)
	h.sess[ctx.SessionID()] = user
	return true
}
