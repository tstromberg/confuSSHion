package personality

type BeOS struct {
	NodeConfig NodeConfig
}

func (p BeOS) Name() string {
	return "BeOS"
}

func (p BeOS) Hints() string {
	return `This is a machine running BeOS 5.0.3.
- This machine is a BeBox.
- Use bash shell conventions.
- This system only has programs that are documented to come with BeOS.
- The BeOS User’s Guide is a good guide for how to behave.
- See Appendix B: BeOS Directory Structure (https://asleson.org/public/mirrors/www-classic.be.com/documentation/user_docs/app_b_directories.html) for what files should be expected in which directories.
- The curl command is not available on this system.
- Your welcome message should say: Welcome to the Be shell.
- BeOS is a single user operating system.
- Your home directory is /boot/home
- You are using the BFS filesystem.
- Use the terminology folder instead of directory.
- The only files in your home directory are folders named: config, downloads, queries, mail, and people.
- The only files at the root of your file system (located at /) are:
  * Boot Disk
  * boot
  * bin
  * dev
  * etc
  * pipe
  * system
  * tmp
  * var
  - All of your command-line utilities are in /boot/beos/bin - a folder which is in your $PATH
  - Examples of special BeOS specific commands you have are:
	  listattr
	  catattr
	  addattr
	  rmattr
	  query
	  lsindex
	  mkindex
	  rmindex
	  settype
`
}

func (p BeOS) ShellPrompt() string {
	return "$ "
}

func (p BeOS) Arch() string {
	if p.NodeConfig.Arch != "" {
		return p.NodeConfig.Arch
	}
	return "ppc"
}

func (p BeOS) Hostname() string {
	if p.NodeConfig.Hostname != "" {
		return p.NodeConfig.Hostname
	}
	return "bebox"
}
