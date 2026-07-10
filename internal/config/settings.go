package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Settings holds the BIFROST_* runtime configuration.
type Settings struct {
	Interface string

	CheckInterval time.Duration
	StaleAfter    time.Duration

	Resolve        bool
	ResolveRetries int
	ResolveBackoff time.Duration

	Reconnect        bool
	ReconnectRetries int
	ReconnectBackoff time.Duration

	Healthcheck      bool
	HealthStaleAfter time.Duration

	Probe         bool
	ProbeInterval time.Duration
	ProbeFails    int
	ProbeTimeout  time.Duration
	ProbeHost     string
}

// LoadSettings reads and validates the BIFROST_* variables via getenv (pass
// os.Getenv in production). It returns the first validation error, if any.
func LoadSettings(getenv func(string) string) (*Settings, error) {
	e := &envReader{getenv: getenv}
	s := &Settings{
		Interface:        e.str("BIFROST_INTERFACE", "wg0"),
		CheckInterval:    e.secs("BIFROST_CHECK_INTERVAL", 30, 1),
		StaleAfter:       e.secs("BIFROST_STALE_AFTER", 135, 0),
		Resolve:          e.boolean("BIFROST_RESOLVE", true),
		ResolveRetries:   e.intMin("BIFROST_RESOLVE_RETRIES", 5, 0),
		ResolveBackoff:   e.secs("BIFROST_RESOLVE_BACKOFF", 5, 0),
		Reconnect:        e.boolean("BIFROST_RECONNECT", true),
		ReconnectRetries: e.intMin("BIFROST_RECONNECT_RETRIES", 5, 0),
		ReconnectBackoff: e.secs("BIFROST_RECONNECT_BACKOFF", 5, 0),
		Healthcheck:      e.boolean("BIFROST_HEALTHCHECK", true),
		HealthStaleAfter: e.secs("BIFROST_HEALTH_STALE_AFTER", 180, 0),
		Probe:            e.boolean("BIFROST_PROBE", false),
		ProbeInterval:    e.secs("BIFROST_PROBE_INTERVAL", 10, 1),
		ProbeFails:       e.intMin("BIFROST_PROBE_FAILS", 3, 1),
		ProbeTimeout:     e.secs("BIFROST_PROBE_TIMEOUT", 2, 1),
		ProbeHost:        e.str("BIFROST_PROBE_HOST", ""),
	}
	if e.err != nil {
		return nil, e.err
	}
	return s, nil
}

type envReader struct {
	getenv func(string) string
	err    error
}

func (e *envReader) str(name, def string) string {
	if v := e.getenv(name); v != "" {
		return v
	}
	return def
}

func (e *envReader) intMin(name string, def, min int) int {
	v := strings.TrimSpace(e.getenv(name))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		e.setErr(fmt.Errorf("%s must be a non-negative integer, got %q", name, v))
		return def
	}
	if n < min {
		e.setErr(fmt.Errorf("%s must be >= %d, got %d", name, min, n))
		return def
	}
	return n
}

func (e *envReader) secs(name string, def, min int) time.Duration {
	return time.Duration(e.intMin(name, def, min)) * time.Second
}

func (e *envReader) boolean(name string, def bool) bool {
	switch strings.ToLower(strings.TrimSpace(e.getenv(name))) {
	case "":
		return def
	case "on", "true", "yes", "1":
		return true
	case "off", "false", "no", "0":
		return false
	default:
		e.setErr(fmt.Errorf("%s must be on/off, got %q", name, e.getenv(name)))
		return def
	}
}

func (e *envReader) setErr(err error) {
	if e.err == nil {
		e.err = err
	}
}
