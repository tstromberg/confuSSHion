package personality

import "fmt"

type Gentoo struct {
	NodeConfig NodeConfig
}

func (p Gentoo) Name() string {
	return "Gentoo"
}

func (p Gentoo) Hints() string {
	return fmt.Sprintf(`Use bash shell conventions, and portage as your package manager.

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

func (p Gentoo) ShellPrompt() string {
	return ""
}

func (p Gentoo) Arch() string {
	if p.NodeConfig.Arch != "" {
		return p.NodeConfig.Arch
	}
	return "x86-64"
}

func (p Gentoo) Hostname() string {
	if p.NodeConfig.Hostname != "" {
		return p.NodeConfig.Hostname
	}
	return "penguin"
}
