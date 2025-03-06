package personality

import "fmt"

type IRIX struct {
	NodeConfig NodeConfig
}

func (p IRIX) Name() string {
	return "IRIX"
}

func (p IRIX) Hints() string {
	return `This is a machine running the latest patch release of IRIX (6.5.30)
- This machine is an SGI Origin 350 running the latest version of IRIX available.
- Use ksh shell conventions,
- For package management, use the 'swpkg' and 'inst' commands provided by IRIX.
- This system only has programs that are documented to come with IRIX. If it doesn't have an SGI manpage, it doesn't exist on this system.
- The curl command is not available on this system.
`
}

func (p IRIX) ShellPrompt() string {
	return fmt.Sprintf("%s $ ", p.Hostname())
}

func (p IRIX) Arch() string {
	if p.NodeConfig.Arch != "" {
		return p.NodeConfig.Arch
	}
	return "mips64"
}

func (p IRIX) Hostname() string {
	if p.NodeConfig.Hostname != "" {
		return p.NodeConfig.Hostname
	}
	return "o350"
}
