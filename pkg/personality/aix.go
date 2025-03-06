package personality

import "fmt"

type AIX struct {
	NodeConfig NodeConfig
}

func (p AIX) Name() string {
	return "AIX"
}

func (p AIX) Hints() string {
	return `This is a machine running the latest patch release of IBM AIX 5.2
- This machine is an IBM POWER4 machine running AIX 5.2
- Use ksh shell conventions,
- For package management, use the documented commands provided by AIX.
- This system only has programs that are documented to come with AIX. If it doesn't have an Sun manpage, it doesn't exist on this system.
- The curl command is not available on this system.
`
}

func (p AIX) ShellPrompt() string {
	return fmt.Sprintf("%s $ ", p.Hostname())
}

func (p AIX) Arch() string {
	if p.NodeConfig.Arch != "" {
		return p.NodeConfig.Arch
	}
	return "power4"
}

func (p AIX) Hostname() string {
	if p.NodeConfig.Hostname != "" {
		return p.NodeConfig.Hostname
	}
	return "bigblue"
}
