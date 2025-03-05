package personality

// NodeConfig represents the server configuration
type NodeConfig struct {
	OS              string
	OSVersion       string
	Arch            string
	IP              string
	Hostname        string
	RoleDescription string
}

// Personality is a common interface for OS personalities
type Personality interface {
	// ShellPrompt returns the OS-specific shell prompt or empty to defer to AI
	ShellPrompt() string
	Hints() string
	Arch() string
	Hostname() string
}

// New returns a new personality for a given environment
func New(nc NodeConfig) Personality {
	switch nc.OS {
	case "openbsd":
		return &OpenBSD{NodeConfig: nc}
	default:
		return &Ubuntu{NodeConfig: nc}
	}
}
