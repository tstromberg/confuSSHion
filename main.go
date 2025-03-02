package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/google/generative-ai-go/genai"
	"github.com/tstromberg/confuSSHion/pkg/auth"
	"github.com/tstromberg/confuSSHion/pkg/holodeck"
	"github.com/tstromberg/confuSSHion/pkg/personality"
	"google.golang.org/api/option"
	"k8s.io/klog/v2"
)

var (
	portFlag            = flag.Int("port", 2222, "Port to listen on for SSH connections")
	promptFlag          = flag.String("prompt", "this machine acts as a firewall and proxy protecting very sensitive data", "Extra prompt for Gemini")
	hostnameFlag        = flag.String("hostname", "", "Custom hostname to use")
	distroFlag          = flag.String("dist", "ubuntu", "Target distribution (aix, arch, fedora, freebsd, hpux, irix, nextstep, openbsd, solaris)")
	githubOrgFlag       = flag.String("github-org", "", "GitHub organization to require users to be part of")
	refreshIntervalFlag = flag.Duration("github-refresh-interval", 12*time.Hour, "Interval to refresh GitHub SSH keys")
	publicKeyAuthFlag   = flag.Bool("public-key-auth", false, "require public key auth")
)

func main() {
	flag.Parse()

	// Initialize Vertex AI client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatalf("Failed to create Vertex AI client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")

	var a auth.Authenticator

	if *githubOrgFlag != "" {
		a = auth.NewGitHubAuthenticator(*githubOrgFlag, 2*time.Hour)
		a.UpdateLoop()
	} else {
		a = auth.NewPermissiveAuthenticator()
	}
	defer a.Close()

	h := holodeck.New(ctx, model, personality.NodeConfig{
		OS:            *distroFlag,
		Hostname:      *hostnameFlag,
		RoleHints:     *promptFlag,
		Authenticator: a,
	})

	// SSH server setup
	ssh.Handle(func(s ssh.Session) {
		h.Handler(s)
	})

	// Start SSH server on specified port
	listenAddr := fmt.Sprintf(":%d", *portFlag)
	log.Printf("Listening on %s as %s honeypot...", listenAddr, *distroFlag)

	if *publicKeyAuthFlag {
		klog.Infof("using public key handler ...")
		err = ssh.ListenAndServe(listenAddr, nil, ssh.PublicKeyAuth(h.PublicKeyHandler))

	} else {
		klog.Infof("using nil auth handler ...")
		err = ssh.ListenAndServe(listenAddr, nil)
	}
	if err != nil {
		klog.Fatalf("listen: %w", err)
	}
}
