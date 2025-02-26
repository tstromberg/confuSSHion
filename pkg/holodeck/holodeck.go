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
	"golang.org/x/crypto/ssh/terminal"
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

func (h Holodeck) Handler(s ssh.Session) error {
	log.Printf("ssh key: %s", s.PublicKey())
	h.p = personality.New(h.nc, personality.UserInfo{
		User:       s.User(),
		RemoteAddr: s.RemoteAddr().String(),
		Environ:    s.Environ(),
		//		PublicKey:  fmt.Sprintf("%s", gossh.MarshalAuthorizedKey(s.PublicKey())),
		Command: s.Command(),
	})

	// Create a pseudo-terminal with distribution-specific prompt
	term := terminal.NewTerminal(s, "")

	basePrompt := h.p.AIPrompt()

	resp, err := h.model.GenerateContent(h.ctx, genai.Text(
		basePrompt+"\n\n"+
			"The user has just logged into the system, welcome them with a standard login message. If your standard login procedure shows the last time the user logged in, they have never logged in before."))
	if err != nil {
		klog.Errorf("geerateContent: %v")
	}
	term.Write([]byte(fmt.Sprintf("%s ", respString(resp))))

	for {
		// Read command from SSH session
		cmd, err := term.ReadLine()
		if err != nil {
			klog.Errorf("readline: %v", err)
			break
		}

		time.Sleep(100 * time.Millisecond)
		out, logout := h.ProcessCmd(cmd)
		term.Write([]byte(strings.TrimSpace(fmt.Sprintf("%s", out))))
		if logout {
			break
		}
	}

	return nil
}

// processCmd simulates the output of a command, returning true if the connection should be kept open.
func (h Holodeck) ProcessCmd(cmd string) (string, bool) {
	klog.Infof("cmd: [%s]", strings.TrimSpace(cmd))

	logout := false
	// if strings.TrimSpace(cmd) == "" {
	// 	continue
	//	}

	h.history = append(h.history, cmd)

	if cmd == "exit" || cmd == "logout" {
		logout = true
	}

	prompt := fmt.Sprintf(`%s

		The user just entered the command %q over an interactive SSH session.
		Previously, they entered these commands in order: %s

		Generate the appropriate output for that command, but also incorporate any state changes that the previous commands may have caused. Do not include a shell prompt in your output.
	`, h.p.AIPrompt(), cmd, strings.Join(h.history, "\n"))

	// Send command to Gemini for processing
	resp, err := h.model.GenerateContent(h.ctx, genai.Text(prompt))
	if err != nil {
		klog.Errorf("error: %v", err)
		return "", false
	}

	return respString(resp), logout
}

func respString(resp *genai.GenerateContentResponse) string {
	return fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0])
}

func (h Holodeck) PublicKeyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	klog.Infof("key provided: [marshal=%s]", gossh.MarshalAuthorizedKey(key))
	return true
}
