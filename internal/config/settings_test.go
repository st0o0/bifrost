package config

import (
	"log/slog"
	"strings"
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
	if s.Metrics || s.MetricsAddr != ":9586" {
		t.Errorf("metrics defaults wrong: Metrics=%v MetricsAddr=%q", s.Metrics, s.MetricsAddr)
	}
	if s.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want info", s.LogLevel)
	}
	if s.LogFormat != "json" {
		t.Errorf("LogFormat = %q, want json", s.LogFormat)
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

func TestLoadSettingsMetrics(t *testing.T) {
	s, err := LoadSettings(env(map[string]string{
		"BIFROST_METRICS": "on", "BIFROST_METRICS_ADDR": "127.0.0.1:2112",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !s.Metrics {
		t.Error("expected Metrics=true")
	}
	if s.MetricsAddr != "127.0.0.1:2112" {
		t.Errorf("MetricsAddr = %q, want 127.0.0.1:2112", s.MetricsAddr)
	}
}

func TestLoadSettingsValidation(t *testing.T) {
	cases := map[string]string{
		"BIFROST_CHECK_INTERVAL":  "0",     // below floor 1
		"BIFROST_PROBE_INTERVAL":  "0",     // below floor 1
		"BIFROST_RESOLVE_RETRIES": "abc",   // non-numeric
		"BIFROST_PROBE":           "maybe", // bad bool
	}
	for k, v := range cases {
		if _, err := LoadSettings(env(map[string]string{k: v})); err == nil {
			t.Errorf("%s=%s: expected error", k, v)
		}
	}
}

func TestLoadSettingsLogLevel(t *testing.T) {
	for _, tc := range []struct {
		val  string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"INFO", slog.LevelInfo},
		{"Warn", slog.LevelWarn},
		{"ERROR", slog.LevelError},
	} {
		s, err := LoadSettings(env(map[string]string{"BIFROST_LOG_LEVEL": tc.val}))
		if err != nil {
			t.Fatalf("BIFROST_LOG_LEVEL=%q: %v", tc.val, err)
		}
		if s.LogLevel != tc.want {
			t.Errorf("BIFROST_LOG_LEVEL=%q: got %v, want %v", tc.val, s.LogLevel, tc.want)
		}
	}
}

func TestLoadSettingsLogLevelInvalid(t *testing.T) {
	_, err := LoadSettings(env(map[string]string{"BIFROST_LOG_LEVEL": "verbose"}))
	if err == nil {
		t.Error("expected error for invalid log level")
	}
}

func TestLoadSettingsLogFormat(t *testing.T) {
	for _, val := range []string{"json", "text", "JSON", "TEXT"} {
		s, err := LoadSettings(env(map[string]string{"BIFROST_LOG_FORMAT": val}))
		if err != nil {
			t.Fatalf("BIFROST_LOG_FORMAT=%q: %v", val, err)
		}
		want := strings.ToLower(val)
		if s.LogFormat != want {
			t.Errorf("BIFROST_LOG_FORMAT=%q: got %q, want %q", val, s.LogFormat, want)
		}
	}
}

func TestLoadSettingsLogFormatInvalid(t *testing.T) {
	_, err := LoadSettings(env(map[string]string{"BIFROST_LOG_FORMAT": "xml"}))
	if err == nil {
		t.Error("expected error for invalid log format")
	}
}
