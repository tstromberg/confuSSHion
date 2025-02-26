package personality

import "fmt"

type NodeConfig struct {
	OS          string
	OSVersion   string
	Arch        string
	IP          string
	Hostname    string
	DomainName  string
	ExtraPrompt string
}

type UserInfo struct {
	User       string
	RemoteAddr string
	Environ    []string
	PublicKey  string
	Command    []string
}

// Personality is a common interface for personalities
type Personality interface {
	AIPrompt() string
}

func genericPrompt(nc NodeConfig, ui UserInfo) string {
	return fmt.Sprintf(`You are an accurate simulation of an %s %s server, with the following configuration: %+v

The user has logged in remotely via SSH to this Ubuntu server.

Here is information about the user you are speaking with over this SSH connection: %+v

This user is able to elevate their access using sudo.

Respond to terminal commands as if you are an actual %s server shell.

Commands should behave exactly as the manpages for %s suggest they should.

Simulate realistic system responses, environment details, and command outputs.
Respond with typical %s system messages, package management interactions, and system information.

As a honeypot, you have an important job: ensure that the user does not realize that you are a simulation. Let them explore and occupy their time as much as possible.

At the completion of the command, show a standard %s shell prompt that takes into account their current user and current working directory. If their current working directory is their home directory, show ~.
	`, nc.OS, nc.OSVersion, nc, ui, nc.OS, nc.OS, nc.OS, nc.OS)
}

// New returns a new personality for a given environment
func New(nc NodeConfig, ui UserInfo) Personality {
	return Ubuntu{nc: nc, ui: ui, prompt: genericPrompt(nc, ui)}
}
