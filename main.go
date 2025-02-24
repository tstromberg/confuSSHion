package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/google/generative-ai-go/genai"
	"github.com/tstromberg/confuSSHion/pkg/personality"
	cssh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"google.golang.org/api/option"
	"k8s.io/klog/v2"
)

var (
	portFlag     = flag.Int("port", 2222, "Port to listen on for SSH connections")
	promptFlag   = flag.String("prompt", "", "Custom prompt for Gemini")
	hostnameFlag = flag.String("hostname", "", "Custom hostname to use")
	distroFlag   = flag.String("dist", "openbsd", "Target distribution (aix, arch, fedora, freebsd, hpux, irix, nextstep, openbsd, solaris)")

	publicKeyAuthFlag = flag.Bool("public-key-auth", false, "require public key auth")
)

func getDistributionPrompt(d string) string {
	switch d {
	case "ubuntu":
		return personality.Ubuntu
	case "openbsd":
		return personality.OpenBSD
	case "solaris":
		return personality.Solaris
	case "arch":
		return personality.Arch
	case "irix":
		return personality.IRIX
	case "hpux":
		return personality.HPUX
	case "gentoo":
		return personality.Gentoo
	case "aix":
		return personality.AIX
	case "netbsd":
		return personality.NetBSD
	case "nextstep":
		return personality.NextSTEP
	}
	return ""
}

func getTerminalPrompt(dist string) string {
	dist = strings.ToLower(dist)
	switch dist {
	case "ubuntu":
		return "ubuntu@ubuntu:~$ "
	case "openbsd":
		return "$ "
	case "solaris":
		return "$ "
	case "arch":
		return "[arch]$ "
	default:
		return "$ "
	}
}

func main() {
	flag.Parse()

	// Determine the effective prompt
	effectivePrompt := *promptFlag
	if effectivePrompt == "" {
		effectivePrompt = getDistributionPrompt(*distroFlag)
	}

	// Initialize Vertex AI client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatalf("Failed to create Vertex AI client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")

	// SSH server setup
	ssh.Handle(func(s ssh.Session) {
		log.Printf("New SSH session from %+v", s)

		// Create a pseudo-terminal with distribution-specific prompt
		term := terminal.NewTerminal(s, getTerminalPrompt(*distroFlag))

		// Initial system context
		initialContext := effectivePrompt

		previous := []string{}

		for {
			// Read command from SSH session
			cmd, err := term.ReadLine()
			if err != nil {
				break
			}

			time.Sleep(100 * time.Millisecond)

			klog.Infof("cmd: [%s]", strings.TrimSpace(cmd))
			if strings.TrimSpace(cmd) == "" {
				continue
			}

			if cmd == "exit" || cmd == "logout" {
				break
			}

			// Combine initial context with current command
			fullPrompt := fmt.Sprintf(`%s

			The user just entered the command %q over an interactive SSH session.
			Previously, they entered these commands in order: %s
			Pretending you are a default %s system, generate the appropriate output for that command, but also incorporate any state changes that the previous commands may have caused. Do not include a shell prompt in your output.
			`, initialContext, cmd, strings.Join(previous, "\n"), *distroFlag)

			previous = append(previous, cmd)
			// Send command to Gemini for processing
			resp, err := model.GenerateContent(ctx, genai.Text(fullPrompt))
			if err != nil {
				term.Write([]byte(fmt.Sprintf("Error: %v\n", err)))
				continue
			}

			// Write Gemini's response back to the SSH session
			rt := resp.Candidates[0].Content.Parts[0]
			rts := strings.ReplaceAll(fmt.Sprintf("%v", rt), "hostname$", "")
			term.Write([]byte(fmt.Sprintf("%s\n", rts)))

			// Update context
			initialContext += fmt.Sprintf("\nPrevious command: %s\nPrevious output: %v", cmd, rt)
		}
	})

	// Start SSH server on specified port
	listenAddr := fmt.Sprintf(":%d", *portFlag)
	log.Printf("Listening on %s as %s honeypot...", listenAddr, *distroFlag)
	if *publicKeyAuthFlag {
		log.Fatal(ssh.ListenAndServe(listenAddr, nil, ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			klog.Infof("key provided: [marshal=%s]", key.Marshal())
			klog.Infof("key provided: [marshal=%s]", cssh.MarshalAuthorizedKey(key))
			return true
		})))
	} else {
		log.Fatal(ssh.ListenAndServe(listenAddr, nil))
	}
}
