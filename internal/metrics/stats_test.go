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
