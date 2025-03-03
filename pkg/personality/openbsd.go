package personality

import "fmt"

type OpenBSD struct {
	nc            NodeConfig
	ui            UserInfo
	cmdPrompt     string
	welcomePrompt string
}

func (p OpenBSD) WelcomePrompt() string {
	return p.welcomePrompt
}

func (p OpenBSD) CommandPrompt() string {
	return fmt.Sprintf(`%s.
Use ksh shell conventions, pkg package manager, and a default OpenBSD system configuration. Assume no system customization, other than a hostname, has been made.

This user does not have permission to switch users or elevate themselves to root without an appropriate exploit.`, p.cmdPrompt)
}

func (p OpenBSD) ShellPrompt() string {
	return fmt.Sprintf("%s $ ", p.nc.Hostname)
}
