package personality

import "fmt"

type Ubuntu struct {
	nc            NodeConfig
	ui            UserInfo
	cmdPrompt     string
	welcomePrompt string
	shellPrompt   string
}

func (p Ubuntu) CommandPrompt() string {
	return fmt.Sprintf(`%s. Use bash shell conventions, apt package manager.
	`, p.cmdPrompt)
}

func (p Ubuntu) WelcomePrompt() string {
	return p.welcomePrompt
}

func (p Ubuntu) ShellPrompt() string {
	// autodetect
	return ""
}
