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
	shellPromptRe = regexp.MustCompile(`\$|\>|\%|\#`)
	templates     *template.Template

	// cacheable defines commands whose output can be cached
	cacheable = map[string]bool{
		"ls": true, "id": true, "w": true, "last": true, "ps": true,
		"clear": true, "cat": true, "echo": true, "less": true, "man": true,
		"whoami": true, "grep": true, "head": true, "tail": true, "diff": true,
		"cmp": true, "comm": true, "sort": true, "df": true, "ifconfig": true,
		"whereis": true, "whatis": true, "apropos": true, "top": true, "pwd": true,
		"wc": true, "find": true, "uptime": true, "du": true, "awk": true,
		"sed": true, "uniq": true, "htop": true, "iostat": true, "vmstat": true,
		"netstat": true, "ss": true, "dig": true, "traceroute": true, "nslookup": true,
		"ip": true, "route": true, "hostname": true, "domainname": true, "lscpu": true,
		"lsblk": true, "lsmod": true, "dmesg": true, "nc": true,
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
