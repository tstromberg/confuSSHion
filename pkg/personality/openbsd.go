package personality

import "fmt"

type OpenBSD struct {
	nc     NodeConfig
	ui     UserInfo
	prompt string
}

func (p OpenBSD) AIPrompt() string {
	return fmt.Sprintf(`%s. Use ksh shell conventions, pkg package manager, and a default OpenBSD system configuration.
	`, p.prompt)
}
