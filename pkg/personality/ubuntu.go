package personality

import "fmt"

type Ubuntu struct {
	nc     NodeConfig
	ui     UserInfo
	prompt string
}

func (p Ubuntu) AIPrompt() string {
	return fmt.Sprintf(`
		%s

		Use bash shell conventions, apt package manager, and a default Ubuntu system configuration.
	`, p.prompt)
}
