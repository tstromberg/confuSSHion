package personality

type OpenVMS struct {
	NodeConfig NodeConfig
}

func (p OpenVMS) Name() string {
	return "OpenVMS"
}

func (p OpenVMS) Hints() string {
	return `This is a machine running the latest patch release of OpenVMS for VAX (7.3)
- This machine is an VAX 7000 running OpenVMS by Digital Equipment Corporation
- You are not running UNIX, you are running OpenVMS
- Use VMS shell conventions,
- This system only has programs that are documented to come with VMS. If it doesn't have an VMS manpage, it doesn't exist on this system.
- This system only understands VMS commands. It does not understand UNIX.
- Instead of the "ls" command to show a file list, OpenVMS uses DIRECTORY
- Instead of the "ps" command to show processes, OpenVMS uses "SHOW SYSTEM"
- OpenVMS commands are always in CAPITAL LETTERS. lowercase commands are translated to uppercase.
`
}

func (p OpenVMS) ShellPrompt() string {
	return "$ "
}

func (p OpenVMS) Arch() string {
	if p.NodeConfig.Arch != "" {
		return p.NodeConfig.Arch
	}
	return "vax"
}

func (p OpenVMS) Hostname() string {
	if p.NodeConfig.Hostname != "" {
		return p.NodeConfig.Hostname
	}
	return "vaxen"
}
