package metrics

import (
	"sync"
	"testing"
)

func TestStatsIncrements(t *testing.T) {
	var s Stats
	s.IncReconnects()
	s.IncReconnects()
	s.IncResolves()
	s.IncEndpointChanges()
	s.IncEndpointChanges()
	s.IncEndpointChanges()

	if got := s.Reconnects(); got != 2 {
		t.Errorf("Reconnects() = %d, want 2", got)
	}
	if got := s.Resolves(); got != 1 {
		t.Errorf("Resolves() = %d, want 1", got)
	}
	if got := s.EndpointChanges(); got != 3 {
		t.Errorf("EndpointChanges() = %d, want 3", got)
	}
}

func TestStatsProbeRTT(t *testing.T) {
	var s Stats
	if got := s.ProbeRTT(); got != 0 {
		t.Errorf("initial ProbeRTT() = %f, want 0", got)
	}
	s.SetProbeRTT(0.012)
	if got := s.ProbeRTT(); got != 0.012 {
		t.Errorf("ProbeRTT() = %f, want 0.012", got)
	}
}

func TestStatsConcurrent(t *testing.T) {
	var s Stats
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)
		go func() { defer wg.Done(); s.IncReconnects() }()
		go func() { defer wg.Done(); s.IncResolves() }()
		go func() { defer wg.Done(); s.IncEndpointChanges() }()
	}
	wg.Wait()
	if got := s.Reconnects(); got != 100 {
		t.Errorf("Reconnects() = %d, want 100", got)
	}
	if got := s.Resolves(); got != 100 {
		t.Errorf("Resolves() = %d, want 100", got)
	}
	if got := s.EndpointChanges(); got != 100 {
		t.Errorf("EndpointChanges() = %d, want 100", got)
	}
}

func TestStatsRecovery(t *testing.T) {
	var s Stats

	if got := s.RecoveryStart(); got != 0 {
		t.Errorf("initial RecoveryStart() = %d, want 0", got)
	}
	if s.RecoveryInProgress() {
		t.Error("initial RecoveryInProgress() = true, want false")
	}

	s.SetRecoveryStart(1720000000)
	if got := s.RecoveryStart(); got != 1720000000 {
		t.Errorf("RecoveryStart() = %d, want 1720000000", got)
	}
	if !s.RecoveryInProgress() {
		t.Error("RecoveryInProgress() = false, want true")
	}

	s.SetRecoveryStart(0)
	if got := s.RecoveryStart(); got != 0 {
		t.Errorf("RecoveryStart() after reset = %d, want 0", got)
	}
	if s.RecoveryInProgress() {
		t.Error("RecoveryInProgress() after reset = true, want false")
	}

	if got := s.LastRecoveryDuration(); got != 0 {
		t.Errorf("initial LastRecoveryDuration() = %f, want 0", got)
	}
	s.SetLastRecoveryDuration(45.5)
	if got := s.LastRecoveryDuration(); got != 45.5 {
		t.Errorf("LastRecoveryDuration() = %f, want 45.5", got)
	}
}

func TestStatsResolveDuration(t *testing.T) {
	var s Stats

	if got := s.LastResolveDuration(); got != 0 {
		t.Errorf("initial LastResolveDuration() = %f, want 0", got)
	}
	s.SetLastResolveDuration(0.25)
	if got := s.LastResolveDuration(); got != 0.25 {
		t.Errorf("LastResolveDuration() = %f, want 0.25", got)
	}
}

func TestStatsTunnelUpSince(t *testing.T) {
	var s Stats

	if got := s.TunnelUpSince(); got != 0 {
		t.Errorf("initial TunnelUpSince() = %d, want 0", got)
	}
	s.SetTunnelUpSince(1720000000)
	if got := s.TunnelUpSince(); got != 1720000000 {
		t.Errorf("TunnelUpSince() = %d, want 1720000000", got)
	}
}
