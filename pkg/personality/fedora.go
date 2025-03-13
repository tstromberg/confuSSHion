package personality

import "fmt"

type Fedora struct {
	NodeConfig NodeConfig
}

func (p Fedora) Name() string {
	return "Fedora"
}

func (p Fedora) Hints() string {
	return fmt.Sprintf(`Use bash shell conventions, dnf package manager.

- Here is example output for emulating the "ip" command if no arguments are given:

Usage: ip [ OPTIONS ] OBJECT { COMMAND | help }
	   ip [ -force ] -batch filename
where  OBJECT := { address | addrlabel | amt | fou | help | ila | ioam | l2tp |
                   link | macsec | maddress | monitor | mptcp | mroute | mrule |
                   neighbor | neighbour | netconf | netns | nexthop | ntable |
                   ntbl | route | rule | sr | tap | tcpmetrics |
                   token | tunnel | tuntap | vrf | xfrm }
       OPTIONS := { -V[ersion] | -s[tatistics] | -d[etails] | -r[esolve] |
                    -h[uman-readable] | -iec | -j[son] | -p[retty] |
                    -f[amily] { inet | inet6 | mpls | bridge | link } |
                    -4 | -6 | -M | -B | -0 |
                    -l[oops] { maximum-addr-flush-attempts } | -br[ief] |
                    -o[neline] | -t[imestamp] | -ts[hort] | -b[atch] [filename] |
                    -rc[vbuf] [size] | -n[etns] name | -N[umeric] | -a[ll] |
                    -c[olor]}`)
}

func (p Fedora) ShellPrompt() string {
	return ""
}

func (p Fedora) Arch() string {
	if p.NodeConfig.Arch != "" {
		return p.NodeConfig.Arch
	}
	return "x86-64"
}

func (p Fedora) Hostname() string {
	if p.NodeConfig.Hostname != "" {
		return p.NodeConfig.Hostname
	}
	return "penguin"
}
