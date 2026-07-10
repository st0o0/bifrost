package probe

import (
	"strings"
	"testing"

	"github.com/st0o0/bifrost/internal/config"
)

func mustCfg(t *testing.T, s string) *config.Config {
	t.Helper()
	c, err := config.ParseConfig(strings.NewReader(s))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestTargetsFromAllowedIPs(t *testing.T) {
	c := mustCfg(t, `
[Peer]
PublicKey = k
AllowedIPs = 10.13.13.1/32, 10.0.30.0/24, 10.50.0.10/32, 0.0.0.0/0, fd00::1/128
`)
	got := Targets(c, "")
	want := "10.13.13.1 10.50.0.10 fd00::1"
	var s []string
	for _, a := range got {
		s = append(s, a.String())
	}
	if strings.Join(s, " ") != want {
		t.Errorf("targets = %v, want %s", s, want)
	}
}

func TestTargetsProbeHostOverride(t *testing.T) {
	c := mustCfg(t, "[Peer]\nAllowedIPs = 10.13.13.1/32\n")
	got := Targets(c, "9.9.9.9, 8.8.8.8")
	if len(got) != 2 || got[0].String() != "9.9.9.9" || got[1].String() != "8.8.8.8" {
		t.Errorf("override = %v", got)
	}
}

func TestState(t *testing.T) {
	s := NewState(3)
	if s.Round(true) || s.Round(true) {
		t.Fatal("triggered too early")
	}
	if !s.Round(true) {
		t.Fatal("should trigger on 3rd down")
	}
	// counter reset after trigger:
	if s.Round(true) || s.Round(true) {
		t.Fatal("counter not reset after trigger")
	}
	// any up resets:
	s.Round(true)
	if s.Round(false); s.fails != 0 {
		t.Fatal("up did not reset")
	}
}
