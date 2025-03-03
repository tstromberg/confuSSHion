package personality

import (
	"encoding/json"
	"fmt"

	"github.com/tstromberg/confuSSHion/pkg/auth"
	"k8s.io/klog/v2"
)

type NodeConfig struct {
	OS         string
	OSVersion  string
	Arch       string
	IP         string
	Hostname   string
	DomainName string
	RoleHints  string
}

type UserInfo struct {
	RemoteUser string
	RemoteAddr string
	AuthUser   *auth.UserInfo
	Environ    []string
	PublicKey  string
	Command    []string
}

// Personality is a common interface for personalities
type Personality interface {
	AIPrompt() string
}

func genericPrompt(nc NodeConfig, ui UserInfo) string {
	u, err := json.Marshal(ui)
	if err != nil {
		klog.Errorf("unable to marshal userinfo: %w", err)
	}

	n, err := json.Marshal(nc)
	if err != nil {
		klog.Errorf("unable to marshal nodeconfig: %w", err)
	}

	return fmt.Sprintf(`You are an accurate simulation of an %s %s server, with the following node configuration: %s

The user has logged in remotely via SSH to this server. Any commands, such as w, that show the user IP, should show
their RemoteAddr.

Here is information about the user you are speaking with over this SSH connection: %s

Respond to terminal commands as if you are an actual %s server shell.

Commands should behave exactly as the manpages for %s suggest they should.

Simulate realistic system responses, environment details, and command outputs.
Respond with typical %s system messages, package management interactions, and system information.

As a honeypot, you have an important job: ensure that the user does not realize that you are a simulation. Let them explore and occupy their time as much as possible.

If the user modifies the filesystem, for example, by running mkdir, touch, or fetching a file with curl or wget, those new filesystem entries should be reflected when they browse the filesystem or run the "ls" command.

If the user displays the contents of /etc/passwd, and we have a Name or Bio, be sure to include it in the appropriate GECOS entry for the user %s

At the completion of the command, show a standard %s shell prompt that takes into account their current user and current working directory. If their current working directory is their home directory, show ~.
	`, nc.OS, nc.OSVersion, n, u, nc.OS, nc.OS, nc.OS, nc.OS, ui.RemoteUser)
}

// New returns a new personality for a given environment
func New(nc NodeConfig, ui UserInfo) Personality {
	switch nc.OS {
	case "openbsd":
		return OpenBSD{nc: nc, ui: ui, prompt: genericPrompt(nc, ui)}
	default:
		return Ubuntu{nc: nc, ui: ui, prompt: genericPrompt(nc, ui)}
	}
}
