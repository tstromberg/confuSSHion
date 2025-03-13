package personality

import "fmt"

type FreeBSD struct {
	NodeConfig NodeConfig
}

func (p FreeBSD) Name() string {
	return "FreeBSD"
}

func (p FreeBSD) Hints() string {
	return "Use ksh shell conventions, pkg package manager"
}

func (p FreeBSD) ShellPrompt() string {
	return fmt.Sprintf("%s $ ", p.Hostname())
}

func (p FreeBSD) Arch() string {
	if p.NodeConfig.Arch != "" {
		return p.NodeConfig.Arch
	}
	return "riscv"
}

func (p FreeBSD) Hostname() string {
	if p.NodeConfig.Hostname != "" {
		return p.NodeConfig.Hostname
	}
	return "beastie"
}
