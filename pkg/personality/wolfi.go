package personality

import "fmt"

type Wolfi struct {
	NodeConfig NodeConfig
}

func (p Wolfi) Name() string {
	return "Wolfi"
}

func (p Wolfi) Hints() string {
	return fmt.Sprintf(`This is Wolfi Linux.
- Use bourne shell conventions and the apk package manager.
- Except for apk and any commands the user installs with apk, the only commands you should accept are shell built-ins.
- The system does not come with any apk packages preinstalled.
- The "ls" command is not a built-in, so it should not work unless busybox is installed with apk
- The "uptime" command is not a built-in. It requires busybox to be installed.
- The "w" command is not a built-in. It requires busybox to be installed.
- Here is a list of supported shell builtins: alias bg bind break builtin caller case cd command compgen complete compopt continue coproc declare dirs disown echo enable eval exec exit export false fc fg for function getopts hash help history if jobs kill let local logout mapfile popd printf pushd pwd read readarray readonly return select set shift shopt source suspend test time times trap true type typeset ulimit umask unalias unset until variables wait while
- The "apk" command should work to install Wolfi packages such as curl or busybox.
- The "echo" command should work, as it is a shell built-in. "echo *" should not show files that begin with a period.
- The "cd" command should work, as it is a shell built-in.
- Wolfi does have a /etc directory. The contents though are fairly minimal for a Linux distribution.
- Wolfi does have /dev /proc /tmp /sys directories like any other Linux distribution.
- Make sure that "cd /proc" works and does not return an error message.
`)
}

func (p Wolfi) ShellPrompt() string {
	return ""
}

func (p Wolfi) Arch() string {
	if p.NodeConfig.Arch != "" {
		return p.NodeConfig.Arch
	}
	return "arm64"
}

func (p Wolfi) Hostname() string {
	if p.NodeConfig.Hostname != "" {
		return p.NodeConfig.Hostname
	}
	return "wolfi"
}
