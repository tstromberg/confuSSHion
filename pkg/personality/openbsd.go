package personality

import "fmt"

type OpenBSD struct {
	nc     NodeConfig
	ui     UserInfo
	prompt string
}

func (p OpenBSD) AIPrompt() string {
	return fmt.Sprintf(`%s.
Use ksh shell conventions, pkg package manager, and a default OpenBSD system configuration. Assume no system customization, other than a hostname, has been made.

This user does not have permission to switch users or elevate themselves to root without an appropriate exploit.`, p.prompt)
}

func (p OpenBSD) ShellPrompt() string {
	return fmt.Sprintf("%s $", p.nc.Hostname)
}
