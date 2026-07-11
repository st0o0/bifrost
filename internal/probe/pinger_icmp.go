//go:build linux

package probe

import (
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

// ICMPPinger sends one privileged ICMP echo per Up call.
type ICMPPinger struct{}

func (ICMPPinger) Up(target string, timeout time.Duration) bool {
	p, err := probing.NewPinger(target)
	if err != nil {
		return false
	}
	p.Count = 1
	p.Timeout = timeout
	p.SetPrivileged(true)
	if err := p.Run(); err != nil {
		return false
	}
	return p.Statistics().PacketsRecv > 0
}
