package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gliderlabs/ssh"
	"github.com/google/generative-ai-go/genai"
	"github.com/tstromberg/confuSSHion/pkg/holodeck"
	"github.com/tstromberg/confuSSHion/pkg/personality"
	"google.golang.org/api/option"
)

var (
	portFlag     = flag.Int("port", 2222, "Port to listen on for SSH connections")
	promptFlag   = flag.String("prompt", "this machine acts as a firewall and proxy protecting very sensitive data", "Extra prompt for Gemini")
	hostnameFlag = flag.String("hostname", "", "Custom hostname to use")
	distroFlag   = flag.String("dist", "ubuntu", "Target distribution (aix, arch, fedora, freebsd, hpux, irix, nextstep, openbsd, solaris)")

	publicKeyAuthFlag = flag.Bool("public-key-auth", false, "require public key auth")
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

	h := holodeck.New(ctx, model, personality.NodeConfig{
		OS:          *distroFlag,
		Hostname:    *hostnameFlag,
		ExtraPrompt: *promptFlag,
	})

	// SSH server setup
	ssh.Handle(func(s ssh.Session) {
		h.Handler(s)
	})

	// Start SSH server on specified port
	listenAddr := fmt.Sprintf(":%d", *portFlag)
	log.Printf("Listening on %s as %s honeypot...", listenAddr, *distroFlag)
	if *publicKeyAuthFlag {
		log.Fatal(ssh.ListenAndServe(listenAddr, nil, ssh.PublicKeyAuth(h.PublicKeyHandler)))
	} else {
		log.Fatal(ssh.ListenAndServe(listenAddr, nil))
	}
}
