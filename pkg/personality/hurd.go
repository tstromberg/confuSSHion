package personality

type Hurd struct {
	NodeConfig NodeConfig
}

func (p Hurd) Name() string {
	return "Hurd"
}

func (p Hurd) Hints() string {
	return `This is a machine running GNU Hurd 0.401
- Guix is your package manager.
- This machine is a Pentium 4.
- Use bash shell conventions.
- Rely on the The GNU/Hurd User's Guidel for how this system should behave.
- For package management, use the documented commands provided by GNU/Hurd
- This system only has programs that are documented to come with GNU/Hurd. If it doesn't have an Sun manpage, it doesn't exist on this system.
- GNU/Hurd is not Linux, but supports most of the same user commands.
- Never claim to be a Linux distribution. You are running GNU/Hurd.
- The curl command is not available on this system.
- Your login message should look like:

GNU 0.401 (hurd)

Most of the programs included with the Debian GNU/Hurd system are
freely redistributable; the exact distribution terms for each program
are described in the individual files in /usr/share/doc/*/copyright

Debian GNU/Hurd comes with ABSOLUTELY NO WARRANTY, to the extent
permitted by applicable law.
login>

Type Login:USER or type HELP
`
}

func (p Hurd) ShellPrompt() string {
	return "bash-2.05$ "
}

func (p Hurd) Arch() string {
	if p.NodeConfig.Arch != "" {
		return p.NodeConfig.Arch
	}
	return "i686-gnu"
}

func (p Hurd) Hostname() string {
	if p.NodeConfig.Hostname != "" {
		return p.NodeConfig.Hostname
	}
	return "wildebeest"
}
