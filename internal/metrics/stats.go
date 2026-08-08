package metrics

import (
	"math"
	"sync/atomic"
)

// Stats holds process-lifetime counters incremented by the supervisor loop
// and read by the Prometheus Collector. All methods are safe for concurrent use.
type Stats struct {
	reconnects      atomic.Uint64
	resolves        atomic.Uint64
	endpointChanges atomic.Uint64
	probeRTT        atomic.Uint64 // float64 bits via math.Float64bits
}

func (s *Stats) IncReconnects()      { s.reconnects.Add(1) }
func (s *Stats) IncResolves()        { s.resolves.Add(1) }
func (s *Stats) IncEndpointChanges() { s.endpointChanges.Add(1) }

func (s *Stats) SetProbeRTT(seconds float64) {
	s.probeRTT.Store(math.Float64bits(seconds))
}

func (s *Stats) Reconnects() uint64      { return s.reconnects.Load() }
func (s *Stats) Resolves() uint64        { return s.resolves.Load() }
func (s *Stats) EndpointChanges() uint64  { return s.endpointChanges.Load() }
func (s *Stats) ProbeRTT() float64       { return math.Float64frombits(s.probeRTT.Load()) }
