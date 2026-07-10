package probe

import (
	"strings"

	"github.com/st0o0/bifrost/internal/config"
)

// Targets returns the probe targets as strings (IPs or, for a PROBE_HOST
// override, possibly hostnames — resolution is the pinger's job). Without an
// override it derives the /32 (IPv4) and /128 (IPv6) host entries of every
// peer's AllowedIPs.
func Targets(cfg *config.Config, probeHost string) []string {
	if strings.TrimSpace(probeHost) != "" {
		return parseHostList(probeHost)
	}
	var out []string
	for _, p := range cfg.Peers {
		for _, ip := range p.AllowedIPs {
			if ip.Bits() == ip.Addr().BitLen() { // /32 for v4, /128 for v6
				out = append(out, ip.Addr().String())
			}
		}
	}
	return out
}

func parseHostList(s string) []string {
	var out []string
	for _, f := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' }) {
		if f = strings.TrimSpace(f); f != "" {
			out = append(out, f)
		}
	}
	return out
}
