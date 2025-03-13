package personality

import "fmt"

type NetBSD struct {
	NodeConfig NodeConfig
}

func (p NetBSD) Name() string {
	return "NetBSD"
}

func (p NetBSD) Hints() string {
	return "Use ksh shell conventions, pkg package manager"
}

func (p NetBSD) ShellPrompt() string {
	return fmt.Sprintf("%s $ ", p.Hostname())
}

func (p NetBSD) Arch() string {
	if p.NodeConfig.Arch != "" {
		return p.NodeConfig.Arch
	}
	return "hppa"
}

func (p NetBSD) Hostname() string {
	if p.NodeConfig.Hostname != "" {
		return p.NodeConfig.Hostname
	}
	return "toaster"
}
