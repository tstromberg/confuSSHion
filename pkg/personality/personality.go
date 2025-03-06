package personality

import "k8s.io/klog/v2"

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
	Name() string
	ShellPrompt() string
	Hints() string
	Arch() string
	Hostname() string
}

// New returns a new personality for a given environment
func New(nc NodeConfig) Personality {
	switch nc.OS {
	case "aix":
		return &AIX{NodeConfig: nc}
	case "hpux":
		return &HPUX{NodeConfig: nc}
	case "irix":
		return &IRIX{NodeConfig: nc}
	case "nextstep":
		return &NeXTStep{NodeConfig: nc}
	case "openbsd":
		return &OpenBSD{NodeConfig: nc}
	case "openvms":
		return &OpenVMS{NodeConfig: nc}
	case "solaris":
		return &Solaris{NodeConfig: nc}
	case "ubuntu":
		return &Ubuntu{NodeConfig: nc}
	case "ultrix":
		return &Ultrix{NodeConfig: nc}
	case "windows":
		return &Windows{NodeConfig: nc}
	case "wolfi":
		return &Wolfi{NodeConfig: nc}
	default:
		klog.Errorf("unknown personality: %q", nc.OS)
		return nil
	}
}
