package main

import (
	"testing"
	"time"
)

func TestHealthy(t *testing.T) {
	now := time.Now()
	if healthy(time.Time{}, now, time.Minute) {
		t.Error("zero handshake should be unhealthy")
	}
	if !healthy(now.Add(-30*time.Second), now, time.Minute) {
		t.Error("30s < 60s should be healthy")
	}
	if healthy(now.Add(-90*time.Second), now, time.Minute) {
		t.Error("90s > 60s should be unhealthy")
	}
}
