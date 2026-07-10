package config

import (
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	const in = `
[Interface]
PrivateKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
Address = 10.13.13.2/32
ListenPort = 51820   # a comment

[Peer]
PublicKey = BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=
Endpoint = vpn.example.com:51820
AllowedIPs = 10.13.13.1/32, 10.50.0.10/32, 0.0.0.0/0
PersistentKeepalive = 25
`
	cfg, err := ParseConfig(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ListenPort != 51820 {
		t.Errorf("ListenPort = %d", cfg.ListenPort)
	}
	if len(cfg.Addresses) != 1 || cfg.Addresses[0].String() != "10.13.13.2/32" {
		t.Errorf("Addresses = %v", cfg.Addresses)
	}
	if len(cfg.Peers) != 1 {
		t.Fatalf("peers = %d", len(cfg.Peers))
	}
	p := cfg.Peers[0]
	if p.EndpointHost != "vpn.example.com" || p.EndpointPort != 51820 {
		t.Errorf("endpoint = %q:%d", p.EndpointHost, p.EndpointPort)
	}
	if len(p.AllowedIPs) != 3 {
		t.Errorf("allowedips = %v", p.AllowedIPs)
	}
	if p.PersistentKeepalive != 25 {
		t.Errorf("keepalive = %d", p.PersistentKeepalive)
	}
}

func TestParseConfigError(t *testing.T) {
	if _, err := ParseConfig(strings.NewReader("PrivateKey = x\n")); err == nil {
		t.Error("expected error for setting outside a section")
	}
}

func TestValidateMissingPrivateKey(t *testing.T) {
	cfg := &Config{
		Peers: []Peer{{PublicKey: "k"}},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing PrivateKey")
	}
}

func TestValidateMissingPeerPublicKey(t *testing.T) {
	cfg := &Config{
		PrivateKey: "k",
		Peers:      []Peer{{}},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for missing peer PublicKey")
	}
}

func TestValidateComplete(t *testing.T) {
	cfg := &Config{
		PrivateKey: "k",
		Peers:      []Peer{{PublicKey: "pk"}},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
