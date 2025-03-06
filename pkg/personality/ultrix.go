package personality

import "fmt"

type Ultrix struct {
	NodeConfig NodeConfig
}

func (p Ultrix) Name() string {
	return "Ultrix"
}

func (p Ultrix) Hints() string {
	return `This is a machine running the latest patch release of Ultrix (4.5)
- This machine is an DECstation 5000 running the latest version of Ultrix available.
- Use ksh shell conventions,
- This system only has programs that are documented to come with Ultrix. If it doesn't have an Ultrix manpage, it doesn't exist on this system.
- The curl command is not available on this system.
- The perl command is preinstalled on this system. It understands perl 4 syntax only.
- Make sure that "cd /etc" works and does not return an error message.
`
}

func (p Ultrix) ShellPrompt() string {
	return fmt.Sprintf("%s $ ", p.Hostname())
}

func (p Ultrix) Arch() string {
	if p.NodeConfig.Arch != "" {
		return p.NodeConfig.Arch
	}
	return "mips64"
}

func (p Ultrix) Hostname() string {
	if p.NodeConfig.Hostname != "" {
		return p.NodeConfig.Hostname
	}
	return "dec5k"
}
