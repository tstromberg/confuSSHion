# confuSSHion

## Overview

confuSSHion is a unique LLM-based SSH honeypot that simulates an interactive terminal session for various operating systems.

## Simulations available

confuSSHion supports simulating the following environments:

- 🐦 AIX
- 🦊 Fedora Linux
- 😈 FreeBSD
- 🧠 Gentoo Linux
- 🦑 HP-UX
- 🌈 IRIX
- 🍎 NeXTSTEP
- 🔷 NetBSD
- 🐡 OpenBSD
- 🖥️ OpenVMS
- 🎩 RHEL (Red Hat Enterprise Linux)
- ☀️ Solaris
- 🐧 Ubuntu Linux
- 🌐 Ultrix
- 🧮 UNICOS
- 🪟 Microsoft Windows
- 🐙 Wolfi Linux

## Features

- 🌐 Supports multiple Unix-like distribution personalities
- 🤖 AI-powered command response generation using Gemini
- 🔒 Configurable SSH server with public key authentication
- 🎭 Dynamic terminal simulation
- 🔍 Session history storage and web UI for browsing captured interactions
- 🔐 GitHub organization-based authentication

## Prerequisites

- Go 1.20+

## Usage

2. Set your Gemini API key:
```bash
export GEMINI_API_KEY=your_api_key_here
```

```bash
go run . [flags]
```

### Flags

- `--port`: SSH server port (default: 2222)
- `--prompt`: Extra prompt for Gemini (default: "this machine acts as a firewall and proxy protecting very sensitive data")
- `--hostname`: Custom hostname to use (default: auto-generated based on distribution)
- `--dist`: Target distribution (default: ubuntu) - Options: aix, hpux, irix, nextstep, openbsd, solaris, ubuntu, ultrix, windows, wolfi
- `--arch`: Target architecture (armd64, amd64, hppa, etc.)
- `--github-org`: GitHub organization to require users to be part of
- `--github-refresh-interval`: Interval to refresh GitHub SSH keys (default: 12h)
- `--public-key-auth`: Require public key authentication
- `--history`: Path to BadgerDB history database (if empty, history is not saved)
- `--http-port`: Port for the web UI (0 to disable, default: 8080)

## Examples

Start a basic Ubuntu honeypot:
```bash
go run main.go
```

Run as an OpenBSD server with custom prompt:
```bash
go run main.go --dist openbsd --prompt "This is a DNS server for acme-corp.com"
```

Run with GitHub organization authentication and session history:
```bash
go run main.go --github-org myorganization --history /path/to/history.db
```

Run with public key authentication and custom port:
```bash
go run main.go --port 2223 --dist wolfi --public-key-auth
```

Access the web UI for session history:
```
http://localhost:8080/
```

## Contributing

Contributions welcome! Please read the contributing guidelines and submit pull requests.
