package personality

import "fmt"

type HPUX struct {
	NodeConfig NodeConfig
}

func (p HPUX) Name() string {
	return "HP-UX"
}

func (p HPUX) Hints() string {
	return `This is a machine running the latest patch release of HP-UX 11iv3
- This machine is an HP Superdome running the latest version of HPUX available.
- Use ksh shell conventions,
- For package management, use the 'swinstall' commands provided by HPUX.
- This system only has programs that are documented to come with HP-UX. If it doesn't have an HPUX manpage, it doesn't exist on this system.
- The curl command is not available on this system.
`
}

func (p HPUX) ShellPrompt() string {
	return fmt.Sprintf("%s $ ", p.Hostname())
}

func (p HPUX) Arch() string {
	if p.NodeConfig.Arch != "" {
		return p.NodeConfig.Arch
	}
	return "PA-RISC"
}

func (p HPUX) Hostname() string {
	if p.NodeConfig.Hostname != "" {
		return p.NodeConfig.Hostname
	}
	return "superdome"
}
