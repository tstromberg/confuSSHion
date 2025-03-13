package holodeck

import (
	"context"
	"embed"
	"log"
	"regexp"
	"text/template"

	"github.com/google/generative-ai-go/genai"
	"github.com/tstromberg/confuSSHion/pkg/auth"
	"github.com/tstromberg/confuSSHion/pkg/history"
	"github.com/tstromberg/confuSSHion/pkg/personality"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

var (
	shellPromptRe = regexp.MustCompile(`(\$|\>|\%|\#) *$`)
	templates     *template.Template

	// locallyCacheable defines commands whose output can be cached
	locallyCacheable = map[string]bool{
		"awk":      true,
		"cat":      true,
		"clear":    true,
		"cmp":      true,
		"comm":     true,
		"df":       true,
		"diff":     true,
		"du":       true,
		"echo":     true,
		"find":     true,
		"grep":     true,
		"head":     true,
		"hostname": true,
		"htop":     true,
		"id":       true,
		"ip":       true,
		"last":     true,
		"less":     true,
		"ls":       true,
		"man":      true,
		"netstat":  true,
		"ps":       true,
		"pwd":      true,
		"sed":      true,
		"sort":     true,
		"ss":       true,
		"tail":     true,
		"top":      true,
		"uniq":     true,
		"uptime":   true,
		"vmstat":   true,
		"w":        true,
		"wc":       true,
		"whoami":   true,
	}

	globallyCacheable = map[string]bool{
		"apropos":    true,
		"df":         true,
		"iostat":     true,
		"dig":        true,
		"dmesg":      true,
		"traceroute": true,
		"domainname": true,
		"ifconfig":   true,
		"ip":         true,
		"lsblk":      true,
		"lscpu":      true,
		"lsmod":      true,
		"lspci":      true,
		"man":        true,
		"nc":         true,
		"nslookup":   true,
		"route":      true,
		"SHOW":       true,
		"uname":      true,
		"whatis":     true,
		"whereis":    true,
	}
)

func init() {
	var err error
	templates, err = template.ParseFS(templateFS, "templates/*.tmpl")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
}

// Response represents the result of a command simulation
type Response struct {
	Output      []byte
	ShellPrompt string
	Logout      bool
}

// Holodeck manages SSH session interactions and LLM responses
type Holodeck struct {
	ctx          context.Context
	nc           personality.NodeConfig
	model        *genai.GenerativeModel
	p            personality.Personality
	auth         auth.Authenticator
	historyStore *history.Store
	authUser     map[string]*auth.UserInfo
	cache        map[string]*Response
}

// New returns a new Holodeck instance
func New(ctx context.Context, model *genai.GenerativeModel, nc personality.NodeConfig, histStore *history.Store, a auth.Authenticator) Holodeck {
	p := personality.New(nc)
	return Holodeck{
		ctx:          ctx,
		nc:           nc,
		model:        model,
		historyStore: histStore,
		auth:         a,
		p:            p,
		authUser:     make(map[string]*auth.UserInfo),
		cache:        make(map[string]*Response),
	}
}
