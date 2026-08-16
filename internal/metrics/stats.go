package metrics

import (
	"math"
	"sync/atomic"
)

// Stats holds process-lifetime counters incremented by the supervisor loop
// and read by the Prometheus Collector. All methods are safe for concurrent use.
type Stats struct {
	reconnects          atomic.Uint64
	resolves            atomic.Uint64
	endpointChanges     atomic.Uint64
	probeRTT            atomic.Uint64 // float64 bits via math.Float64bits
	recoveryStart       atomic.Int64  // unix nano; 0 = not in progress
	lastRecoveryDuration atomic.Uint64 // float64 bits (seconds)
	lastResolveDuration  atomic.Uint64 // float64 bits (seconds)
	tunnelUpSince        atomic.Int64  // unix seconds
}

func (s *Stats) IncReconnects()      { s.reconnects.Add(1) }
func (s *Stats) IncResolves()        { s.resolves.Add(1) }
func (s *Stats) IncEndpointChanges() { s.endpointChanges.Add(1) }

func (s *Stats) SetProbeRTT(seconds float64) {
	s.probeRTT.Store(math.Float64bits(seconds))
}

func (s *Stats) Reconnects() uint64     { return s.reconnects.Load() }
func (s *Stats) Resolves() uint64       { return s.resolves.Load() }
func (s *Stats) EndpointChanges() uint64 { return s.endpointChanges.Load() }
func (s *Stats) ProbeRTT() float64      { return math.Float64frombits(s.probeRTT.Load()) }

func (s *Stats) SetRecoveryStart(nanos int64) { s.recoveryStart.Store(nanos) }
func (s *Stats) RecoveryStart() int64          { return s.recoveryStart.Load() }
func (s *Stats) RecoveryInProgress() bool      { return s.recoveryStart.Load() != 0 }

func (s *Stats) SetLastRecoveryDuration(seconds float64) {
	s.lastRecoveryDuration.Store(math.Float64bits(seconds))
}
func (s *Stats) LastRecoveryDuration() float64 {
	return math.Float64frombits(s.lastRecoveryDuration.Load())
}

func (s *Stats) SetLastResolveDuration(seconds float64) {
	s.lastResolveDuration.Store(math.Float64bits(seconds))
}
func (s *Stats) LastResolveDuration() float64 {
	return math.Float64frombits(s.lastResolveDuration.Load())
}

func (s *Stats) SetTunnelUpSince(unix int64) { s.tunnelUpSince.Store(unix) }
func (s *Stats) TunnelUpSince() int64        { return s.tunnelUpSince.Load() }
