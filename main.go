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
	"github.com/tstromberg/confuSSHion/pkg/history"
	"github.com/tstromberg/confuSSHion/pkg/holodeck"
	"github.com/tstromberg/confuSSHion/pkg/personality"
	"github.com/tstromberg/confuSSHion/pkg/ui"
	"google.golang.org/api/option"
	"k8s.io/klog/v2"
)

var (
	portFlag            = flag.Int("port", 2222, "Port to listen on for SSH connections")
	promptFlag          = flag.String("prompt", "this machine acts as a firewall and proxy protecting very sensitive data", "Extra prompt for Gemini")
	hostnameFlag        = flag.String("hostname", "", "Custom hostname to use")
	distroFlag          = flag.String("dist", "ubuntu", "Target distribution (aix, hpux, irix, nextstep, openbsd, solaris, ubuntu, ultrix, windows, wolfi)")
	archFlag            = flag.String("arch", "", "Target architecture (armd64, amd64, hppa, etc)")
	githubOrgFlag       = flag.String("github-org", "", "GitHub organization to require users to be part of")
	refreshIntervalFlag = flag.Duration("github-refresh-interval", 12*time.Hour, "Interval to refresh GitHub SSH keys")
	publicKeyAuthFlag   = flag.Bool("public-key-auth", false, "require public key auth")
	historyPathFlag     = flag.String("history", "", "Path to BadgerDB history database (if empty, history is not saved)")
	httpPortFlag        = flag.Int("http-port", 8080, "Port for the web UI (0 to disable)")
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

	// Initialize history store if path is provided
	var histStore *history.Store
	if *historyPathFlag != "" {
		var err error
		histStore, err = history.NewStore(*historyPathFlag)
		if err != nil {
			log.Fatalf("Failed to create history store: %v", err)
		}
		defer histStore.Close()
		log.Printf("Session history will be saved to: %s", *historyPathFlag)
	} else {
		log.Printf("Session history will not be saved (use --history to enable)")
	}

	if *historyPathFlag != "" && *httpPortFlag > 0 {
		uiServer, err := ui.NewServer(histStore, *httpPortFlag)
		if err != nil {
			log.Fatalf("Failed to create UI server: %v", err)
		}

		go func() {
			if err := uiServer.Start(); err != nil {
				log.Printf("Web UI error: %v", err)
			}
		}()
	} else if *httpPortFlag > 0 {
		log.Printf("Web UI disabled because history storage is not enabled")
	}

	h := holodeck.New(ctx, model, personality.NodeConfig{
		OS:              *distroFlag,
		Hostname:        *hostnameFlag,
		Arch:            *archFlag,
		RoleDescription: *promptFlag,
	}, histStore, a)

	// SSH server setup
	ssh.Handle(func(s ssh.Session) {
		err := h.Handler(s)
		klog.Errorf("handler error: %v", err)
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
