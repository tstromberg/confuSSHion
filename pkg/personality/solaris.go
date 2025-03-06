package personality

import "fmt"

type Solaris struct {
	NodeConfig NodeConfig
}

func (p Solaris) Name() string {
	return "Solaris"
}

func (p Solaris) Hints() string {
	return `This is a machine running the latest patch release of Solaris 10
- This machine is an Sun Fire 15K (Starkitty) running the latest version of Solaris 10 available.
- Use ksh shell conventions,
- For package management, use the documented commands provided by Solaris.
- This system only has programs that are documented to come with Solaris. If it doesn't have an Sun manpage, it doesn't exist on this system.
- The curl command is not available on this system.
`
}

func (p Solaris) ShellPrompt() string {
	return fmt.Sprintf("%s $ ", p.Hostname())
}

func (p Solaris) Arch() string {
	if p.NodeConfig.Arch != "" {
		return p.NodeConfig.Arch
	}
	return "sparc64"
}

func (p Solaris) Hostname() string {
	if p.NodeConfig.Hostname != "" {
		return p.NodeConfig.Hostname
	}
	return "bigsun"
}
