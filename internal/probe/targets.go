package probe

import (
	"net/netip"
	"strings"

	"github.com/st0o0/bifrost/internal/config"
)

// Targets returns the probe targets: probeHost (comma/space list) if non-empty,
// else the /32 (IPv4) and /128 (IPv6) host entries of every peer's AllowedIPs.
func Targets(cfg *config.Config, probeHost string) []netip.Addr {
	if strings.TrimSpace(probeHost) != "" {
		return parseHostList(probeHost)
	}
	var out []netip.Addr
	for _, p := range cfg.Peers {
		for _, ip := range p.AllowedIPs {
			if ip.Bits() == ip.Addr().BitLen() { // /32 for v4, /128 for v6
				out = append(out, ip.Addr())
			}
		}
	}
	return out
}

func parseHostList(s string) []netip.Addr {
	var out []netip.Addr
	fields := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' })
	for _, f := range fields {
		if a, err := netip.ParseAddr(strings.TrimSpace(f)); err == nil {
			out = append(out, a)
		}
	}
	return out
}
