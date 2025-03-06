package personality

import "fmt"

type NeXTStep struct {
	NodeConfig NodeConfig
}

func (p NeXTStep) Name() string {
	return "NeXTStep"
}

func (p NeXTStep) Hints() string {
	return `This is a machine running the latest patch release of NeXTStep (3.3)
- This machine is an NeXTStation Color
- Use ksh shell conventions,
- This system only has programs that are documented to come with NeXTStep. If it doesn't have an NeXTStep manpage, it doesn't exist on this system.
- The curl command is not available on this system.
- The perl command is preinstalled on this system. It understands perl 4 syntax only.
- Make sure that "cd /etc" works and does not return an error message.
`
}

func (p NeXTStep) ShellPrompt() string {
	return fmt.Sprintf("%s $ ", p.Hostname())
}

func (p NeXTStep) Arch() string {
	if p.NodeConfig.Arch != "" {
		return p.NodeConfig.Arch
	}
	return "m68k"
}

func (p NeXTStep) Hostname() string {
	if p.NodeConfig.Hostname != "" {
		return p.NodeConfig.Hostname
	}
	return "slab"
}
