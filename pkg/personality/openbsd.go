package personality

import "fmt"

type OpenBSD struct {
	NodeConfig NodeConfig
}

func (p OpenBSD) Hints() string {
	return "Use ksh shell conventions, pkg package manager"
}

func (p OpenBSD) ShellPrompt() string {
	return fmt.Sprintf("%s $ ", p.Hostname())
}

func (p OpenBSD) Arch() string {
	if p.NodeConfig.Arch != "" {
		return p.NodeConfig.Arch
	}
	return "sparc64"
}

func (p OpenBSD) Hostname() string {
	if p.NodeConfig.Hostname != "" {
		return p.NodeConfig.Hostname
	}
	return "fugu"
}
