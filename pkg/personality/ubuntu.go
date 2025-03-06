package personality

import "fmt"

type Ubuntu struct {
	NodeConfig NodeConfig
}

func (p Ubuntu) Name() string {
	return "Ubuntu"
}

func (p Ubuntu) Hints() string {
	return fmt.Sprintf(`Use bash shell conventions, apt package manager.`)
}

func (p Ubuntu) ShellPrompt() string {
	return ""
}

func (p Ubuntu) Arch() string {
	if p.NodeConfig.Arch != "" {
		return p.NodeConfig.Arch
	}
	return "riscv"
}

func (p Ubuntu) Hostname() string {
	if p.NodeConfig.Hostname != "" {
		return p.NodeConfig.Hostname
	}
	return "ubuntu"
}
