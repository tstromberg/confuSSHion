# confuSSHion

## Overview

confuSSHion is an LLM-based SSH honeypot that simulates an interactive terminal session for a variety of operating systems.

## Available Simulations

confuSSHion simulates many esoteric environments:

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

- 🌐 Multiple Unix-like and Windows personalities
- 🤖 AI-powered responses using Gemini
- 🔒 Configurable SSH server with public key authentication
- 🎭 Dynamic terminal simulation
- 🔍 Session history storage with web UI for reviewing captured interactions
- 🔐 Optional GitHub organization-based authentication

## Prerequisites

- Go 1.20+

## Setup and Usage

1. Set your Gemini API key:
   ```bash
   export GEMINI_API_KEY=your_api_key_here
   ```

2. Run the honeypot:
   ```bash
   go run . [flags]
   ```

### Command-line Options

| Flag | Description | Default |
|------|-------------|---------|
| `--port` | SSH server port | 2222 |
| `--prompt` | Custom prompt for Gemini | "this machine acts as a firewall and proxy protecting very sensitive data" |
| `--hostname` | Custom hostname | auto-generated based on distribution |
| `--dist` | Target distribution | ubuntu |
| `--arch` | Target architecture | - |
| `--github-org` | GitHub organization for authentication | - |
| `--github-refresh-interval` | GitHub SSH key refresh interval | 12h |
| `--public-key-auth` | Enable public key authentication | false |
| `--history` | Path to history database | - |
| `--http-port` | Web UI port (0 to disable) | 8080 |

Available distributions: aix, fedora, freebsd, gentoo, hpux, irix, nextstep, netbsd, openbsd, openvms, rhel, solaris, ubuntu, ultrix, unicos, windows, wolfi

## Examples

**Basic Ubuntu honeypot:**
```bash
go run main.go
```

**OpenBSD DNS server simulation:**
```bash
go run main.go --dist openbsd --prompt "This is a DNS server for acme-corp.com"
```

**GitHub authentication with session history:**
```bash
go run main.go --github-org myorganization --history /path/to/history.db
```

**Custom port with public key authentication:**
```bash
go run main.go --port 2223 --dist wolfi --public-key-auth
```

**Access the web UI for session history:**
```
http://localhost:8080/
```

## Contributing

PR's welcome!
