package personality

type Windows struct {
	NodeConfig NodeConfig
}

func (p Windows) Name() string {
	return "Microsoft Windows NT"
}

func (p Windows) Hints() string {
	return `This is a machine running the latest patch release of Windows NT 4.0
- This machine is an Pentium Pro running Windows NT 4.0 by Microsoft
- You are not running UNIX, you are running Windows
- Use Windows shell conventions
- This system only has programs that are documented to come with Microsoft Windows NT 4.0. If it doesn't have an NT help page, it doesn't exist on this system.
- This system only understands Windows and MS DOS commands. It does not understand UNIX.
- Instead of the "ls" command to show a file list, Windows uses DIR
- This system does not have PowerShell
`
}

func (p Windows) ShellPrompt() string {
	return ""
}

func (p Windows) Arch() string {
	if p.NodeConfig.Arch != "" {
		return p.NodeConfig.Arch
	}
	return "x86"
}

func (p Windows) Hostname() string {
	if p.NodeConfig.Hostname != "" {
		return p.NodeConfig.Hostname
	}
	return "closet"
}
