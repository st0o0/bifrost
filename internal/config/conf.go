package config

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/netip"
	"strconv"
	"strings"
)

// Config is a parsed WireGuard configuration file.
type Config struct {
	PrivateKey string
	Addresses  []netip.Prefix
	ListenPort int
	MTU        int
	Peers      []Peer
}

// Peer is one [Peer] section. EndpointHost is kept verbatim (may be a hostname)
// so it can be re-resolved for DDNS.
type Peer struct {
	PublicKey           string
	PresharedKey        string
	EndpointHost        string
	EndpointPort        int
	AllowedIPs          []netip.Prefix
	PersistentKeepalive int
}

// ParseConfig parses a wg-quick style .conf.
func ParseConfig(r io.Reader) (*Config, error) {
	cfg := &Config{}
	var section string
	var cur *Peer

	sc := bufio.NewScanner(r)
	for line := 1; sc.Scan(); line++ {
		s := sc.Text()
		if i := strings.IndexByte(s, '#'); i >= 0 {
			s = s[:i]
		}
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
			section = strings.ToLower(strings.TrimSpace(s[1 : len(s)-1]))
			if section == "peer" {
				cfg.Peers = append(cfg.Peers, Peer{})
				cur = &cfg.Peers[len(cfg.Peers)-1]
			}
			continue
		}
		key, val, ok := strings.Cut(s, "=")
		if !ok {
			return nil, fmt.Errorf("line %d: expected key = value", line)
		}
		key = strings.TrimSpace(strings.ToLower(key))
		val = strings.TrimSpace(val)

		switch section {
		case "interface":
			if err := cfg.setInterface(key, val); err != nil {
				return nil, fmt.Errorf("line %d: %w", line, err)
			}
		case "peer":
			if err := cur.set(key, val); err != nil {
				return nil, fmt.Errorf("line %d: %w", line, err)
			}
		default:
			if section == "" {
				return nil, fmt.Errorf("line %d: setting outside a section", line)
			}
			return nil, fmt.Errorf("line %d: unknown section %q", line, section)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate checks required fields.
func (c *Config) Validate() error {
	if c.PrivateKey == "" {
		return fmt.Errorf("[Interface] PrivateKey is required")
	}
	for i, p := range c.Peers {
		if p.PublicKey == "" {
			return fmt.Errorf("[Peer] %d: PublicKey is required", i+1)
		}
	}
	return nil
}

func (c *Config) setInterface(key, val string) error {
	switch key {
	case "privatekey":
		c.PrivateKey = val
	case "address":
		ps, err := parsePrefixList(val)
		if err != nil {
			return err
		}
		c.Addresses = append(c.Addresses, ps...)
	case "listenport":
		n, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("ListenPort: %w", err)
		}
		c.ListenPort = n
	case "mtu":
		n, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("MTU: %w", err)
		}
		c.MTU = n
	case "dns":
		// v1: ignored (the caller warn-logs); split-tunnel rarely needs it.
	}
	return nil
}

func (p *Peer) set(key, val string) error {
	switch key {
	case "publickey":
		p.PublicKey = val
	case "presharedkey":
		p.PresharedKey = val
	case "endpoint":
		host, port, err := splitEndpoint(val)
		if err != nil {
			return err
		}
		p.EndpointHost, p.EndpointPort = host, port
	case "allowedips":
		ps, err := parsePrefixList(val)
		if err != nil {
			return err
		}
		p.AllowedIPs = append(p.AllowedIPs, ps...)
	case "persistentkeepalive":
		n, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("PersistentKeepalive: %w", err)
		}
		p.PersistentKeepalive = n
	}
	return nil
}

func parsePrefixList(val string) ([]netip.Prefix, error) {
	var out []netip.Prefix
	for _, part := range strings.Split(val, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		p, err := netip.ParsePrefix(part)
		if err != nil {
			addr, aerr := netip.ParseAddr(part)
			if aerr != nil {
				return nil, fmt.Errorf("invalid CIDR/IP %q: %w", part, err)
			}
			p = netip.PrefixFrom(addr, addr.BitLen())
		}
		out = append(out, p)
	}
	return out, nil
}

func splitEndpoint(val string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(val)
	if err != nil {
		return "", 0, fmt.Errorf("Endpoint %q: %w", val, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("Endpoint port %q: %w", portStr, err)
	}
	return host, port, nil
}
