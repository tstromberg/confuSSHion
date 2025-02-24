# ConfuSSHion

## Overview

ConfuSSHion is an intelligent SSH honeypot that uses AI to simulate interactive terminal sessions for various Unix-like operating systems. By leveraging the Gemini AI model, it generates contextually appropriate responses to user commands, creating a realistic and dynamic honeypot environment.

## Supported Distributions

ConfuSSHion currently supports the following Unix-like distributions:

- 🐧 Ubuntu
- 🐡 OpenBSD
- ☀️ Solaris
- 🏹 Arch Linux
- 🌈 IRIX
- 🖥️ HP-UX
- 🐉 Gentoo
- 🌐 AIX
- 🌊 NetBSD
- ⏳ NextSTEP

## Features

- 🌐 Supports multiple Unix-like distribution personalities
- 🤖 AI-powered command response generation
- 🔒 Configurable SSH server
- 🎭 Dynamic terminal simulation

## Prerequisites

- Go 1.20+
- Gemini API Key
- Required Go dependencies (see `go.mod`)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/confuSSHion.git
cd confuSSHion
```

3. Set your Gemini API key:
```bash
export GEMINI_API_KEY=your_api_key_here
```

## Usage

```bash
go run . [flags]
```

### Flags

- `--port`: SSH server port (default: 2222)
- `--dist`: Target distribution (default: openbsd)
- `--prompt`: Custom AI prompt
- `--public-key-auth`: Require public key authentication

## Example

```bash
go run main.go --port 2223 --dist ubuntu --public-key-auth
```

## Contributing

Contributions welcome! Please read the contributing guidelines and submit pull requests.
