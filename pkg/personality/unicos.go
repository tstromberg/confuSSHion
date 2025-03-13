package personality

import "fmt"

type UNICOS struct {
	NodeConfig NodeConfig
}

func (p UNICOS) Name() string {
	return "UNICOS"
}

func (p UNICOS) Hints() string {
	return `This is a machine running the latest patch release of UNICOS/mk
- This machine is an Cray T3E running the UNICOS/mk 10.0.0
- Use ksh shell conventions.
- UNICOS supports all standard AT&T SYSV UNIX commands, like "ls"
- This system only has programs that are documented to come with UNICOS/mk. If it doesn't have an UNICOS/mk manpage, it doesn't exist on this system.
- Use the UNICOS® Administrator Commands Reference Manual and UNICOS Basic Administrat/eion Guide as a reference for how to behave.
- The curl command is not available on this system.
- The perl command is preinstalled on this system. It understands perl 4 syntax only.
- Make sure that "cd /etc" works and does not return an error message.
- /etc should have standard SYSV UNIX files, but also some UNICOS specific files like /etc/
- You have a program named /etc/nu that is a prompt-driven utility for interactively adding, deleting, and modifying user records.
`
}

func (p UNICOS) ShellPrompt() string {
	return fmt.Sprintf("%s $ ", p.Hostname())
}

func (p UNICOS) Arch() string {
	if p.NodeConfig.Arch != "" {
		return p.NodeConfig.Arch
	}
	return "crayt3e"
}

func (p UNICOS) Hostname() string {
	if p.NodeConfig.Hostname != "" {
		return p.NodeConfig.Hostname
	}
	return "bigcray"
}
