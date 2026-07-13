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

func TestLoadConfigFromEnv(t *testing.T) {
	env := map[string]string{
		"BIFROST_PRIVATE_KEY":       "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"BIFROST_ADDRESS":           "10.13.13.2/32",
		"BIFROST_LISTEN_PORT":       "51820",
		"BIFROST_MTU":               "1420",
		"BIFROST_PEER_PUBLIC_KEY":   "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=",
		"BIFROST_PEER_ENDPOINT":     "vpn.example.com:51820",
		"BIFROST_PEER_ALLOWED_IPS":  "10.13.13.1/32, 10.50.0.10/32",
		"BIFROST_PEER_KEEPALIVE":    "25",
		"BIFROST_PEER_PRESHARED_KEY": "CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC=",
	}
	getenv := func(k string) string { return env[k] }

	cfg, ok, err := LoadConfig(getenv)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected ok=true")
	}
	if cfg.PrivateKey != env["BIFROST_PRIVATE_KEY"] {
		t.Errorf("PrivateKey = %q", cfg.PrivateKey)
	}
	if cfg.ListenPort != 51820 {
		t.Errorf("ListenPort = %d", cfg.ListenPort)
	}
	if cfg.MTU != 1420 {
		t.Errorf("MTU = %d", cfg.MTU)
	}
	if len(cfg.Addresses) != 1 || cfg.Addresses[0].String() != "10.13.13.2/32" {
		t.Errorf("Addresses = %v", cfg.Addresses)
	}
	if len(cfg.Peers) != 1 {
		t.Fatalf("peers = %d", len(cfg.Peers))
	}
	p := cfg.Peers[0]
	if p.PublicKey != env["BIFROST_PEER_PUBLIC_KEY"] {
		t.Errorf("PublicKey = %q", p.PublicKey)
	}
	if p.PresharedKey != env["BIFROST_PEER_PRESHARED_KEY"] {
		t.Errorf("PresharedKey = %q", p.PresharedKey)
	}
	if p.EndpointHost != "vpn.example.com" || p.EndpointPort != 51820 {
		t.Errorf("endpoint = %q:%d", p.EndpointHost, p.EndpointPort)
	}
	if len(p.AllowedIPs) != 2 {
		t.Errorf("allowedips = %v", p.AllowedIPs)
	}
	if p.PersistentKeepalive != 25 {
		t.Errorf("keepalive = %d", p.PersistentKeepalive)
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("validate: %v", err)
	}
}

func TestLoadConfigNoEnv(t *testing.T) {
	getenv := func(string) string { return "" }
	cfg, ok, err := LoadConfig(getenv)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expected ok=false when BIFROST_PRIVATE_KEY is not set")
	}
	if cfg != nil {
		t.Error("expected nil config")
	}
}

func TestLoadConfigMinimal(t *testing.T) {
	env := map[string]string{
		"BIFROST_PRIVATE_KEY":     "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"BIFROST_PEER_PUBLIC_KEY": "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=",
	}
	getenv := func(k string) string { return env[k] }

	cfg, ok, err := LoadConfig(getenv)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected ok=true")
	}
	if cfg.ListenPort != 0 {
		t.Errorf("ListenPort = %d, want 0", cfg.ListenPort)
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("validate: %v", err)
	}
}

func TestLoadConfigBadEndpoint(t *testing.T) {
	env := map[string]string{
		"BIFROST_PRIVATE_KEY":   "k",
		"BIFROST_PEER_ENDPOINT": "not-a-valid-endpoint",
	}
	getenv := func(k string) string { return env[k] }

	_, _, err := LoadConfig(getenv)
	if err == nil {
		t.Error("expected error for bad endpoint")
	}
}

func TestLoadConfigBadAddress(t *testing.T) {
	env := map[string]string{
		"BIFROST_PRIVATE_KEY": "k",
		"BIFROST_ADDRESS":     "not-cidr",
	}
	getenv := func(k string) string { return env[k] }

	_, _, err := LoadConfig(getenv)
	if err == nil {
		t.Error("expected error for bad address")
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
