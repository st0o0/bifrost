package config

import (
	"testing"
	"time"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadSettingsDefaults(t *testing.T) {
	s, err := LoadSettings(env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if s.Interface != "wg0" || s.CheckInterval != 30*time.Second || s.StaleAfter != 135*time.Second {
		t.Errorf("defaults wrong: %+v", s)
	}
	if !s.Resolve || !s.Reconnect || !s.Healthcheck || s.Probe {
		t.Errorf("toggle defaults wrong: %+v", s)
	}
	if s.ProbeInterval != 10*time.Second || s.ProbeFails != 3 {
		t.Errorf("probe defaults wrong: %+v", s)
	}
}

func TestLoadSettingsOverrideAndBool(t *testing.T) {
	s, err := LoadSettings(env(map[string]string{
		"BIFROST_PROBE": "on", "BIFROST_RESOLVE": "off", "BIFROST_STALE_AFTER": "60",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !s.Probe || s.Resolve || s.StaleAfter != 60*time.Second {
		t.Errorf("override wrong: %+v", s)
	}
}

func TestLoadSettingsValidation(t *testing.T) {
	cases := map[string]string{
		"BIFROST_CHECK_INTERVAL": "0",   // below floor 1
		"BIFROST_PROBE_INTERVAL": "0",   // below floor 1
		"BIFROST_RESOLVE_RETRIES": "abc", // non-numeric
		"BIFROST_PROBE":          "maybe", // bad bool
	}
	for k, v := range cases {
		if _, err := LoadSettings(env(map[string]string{k: v})); err == nil {
			t.Errorf("%s=%s: expected error", k, v)
		}
	}
}
