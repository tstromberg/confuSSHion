package holodeck

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"

	"github.com/gliderlabs/ssh"
	"github.com/tstromberg/confuSSHion/pkg/personality"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"k8s.io/klog/v2"
)

// New returns a new holodeck
func New(ctx context.Context, model *genai.GenerativeModel, nc personality.NodeConfig) Holodeck {
	return Holodeck{
		nc:      nc,
		model:   model,
		ctx:     ctx,
		history: []string{},
	}
}

type Holodeck struct {
	// ctx shouldn't be here, but the alternative approaches suck
	ctx     context.Context
	nc      personality.NodeConfig
	model   *genai.GenerativeModel
	history []string
	p       personality.Personality
}

type Response struct {
	Output      []byte
	ShellPrompt string
	Logout      bool
}

func (h Holodeck) Handler(s ssh.Session) error {
	log.Printf("ssh key: %s", s.PublicKey())
	h.p = personality.New(h.nc, personality.UserInfo{
		User:       s.User(),
		RemoteAddr: s.RemoteAddr().String(),
		Environ:    s.Environ(),
		//		PublicKey:  fmt.Sprintf("%s", gossh.MarshalAuthorizedKey(s.PublicKey())),
		Command: s.Command(),
	})

	resp, err := h.simulate("The user has just logged into the system, welcome them with a standard login message and an appropriate shell prompt. If your standard login procedure shows the last time the user logged in, they have never logged in before.")
	if err != nil {
		return err
	}

	time.Sleep(500 * time.Millisecond)
	term := term.NewTerminal(s, resp.ShellPrompt)
	term.Write(resp.Output)
	history := []string{}

	for {
		// Read command from SSH session
		klog.Infof("waiting for input ...")
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

		prompt := fmt.Sprintf(`
			The user just entered the command %q over an interactive SSH session.
			Previously, they entered these commands in order: %s

			Unless the user has changed directories via 'cd', assume their current working directory is their home directory.

			Generate the appropriate output for that command, but also incorporate any state changes that the previous commands may have caused. Do not include a shell prompt in your output.
		`, cmd, strings.Join(h.history, "\n"))

		resp, err := h.simulate(prompt)
		term.Write(resp.Output)

		if cmd == "exit" || cmd == "logout" {
			break
		}

		if resp.ShellPrompt != "" {
			klog.Infof("setting prompt to: %q", resp.ShellPrompt)
			term.SetPrompt(resp.ShellPrompt)
		}
		history = append(history, cmd)
	}

	return nil
}

func (h Holodeck) simulate(prompt string) (*Response, error) {
	fullPrompt := h.p.AIPrompt() + "\n\n" + prompt
	klog.Infof("sending prompt: %s", fullPrompt)
	resp, err := h.model.GenerateContent(h.ctx, genai.Text(fullPrompt))
	if err != nil {
		return nil, err
	}

	output := []string{}
	shellPrompt := ""
	raw := fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0])

	lines := strings.Split(raw, "\n")
	for x, l := range lines {
		// vertex bug workaround, sometimes it returns markdown
		if l == "```" {
			log.Printf("skipping markdown line: %q", l)
			continue
		}
		if x == len(lines)-2 {
			output = append(output, "")
			shellPrompt = l
			break
		}
		output = append(output, l)
	}

	out := strings.Join(output, "\n")
	klog.Infof("vertex response: %s", out)
	klog.Infof("vertex prompt: %s", shellPrompt)

	if !strings.HasSuffix(shellPrompt, " ") {
		shellPrompt = shellPrompt + " "
	}

	return &Response{
		Output:      []byte(out),
		ShellPrompt: shellPrompt,
	}, nil
}

func (h Holodeck) PublicKeyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	klog.Infof("key provided: [marshal=%s]", gossh.MarshalAuthorizedKey(key))
	return true
}
