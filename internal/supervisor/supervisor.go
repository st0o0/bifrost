// Package supervisor drives recovery from two independent triggers: the
// liveness probe (all targets down for N consecutive rounds) and
// handshake-age staleness. It depends only on the recovery.Controller and
// probe.Pinger interfaces plus Phase 1 packages, so it is cross-platform and
// unit-testable without a real WireGuard interface.
package supervisor

import (
	"context"
	"log"
	"time"

	"github.com/st0o0/bifrost/internal/config"
	"github.com/st0o0/bifrost/internal/probe"
	"github.com/st0o0/bifrost/internal/recovery"
)

// Deps bundles the supervisor's collaborators and injectable timing sources.
type Deps struct {
	Ctrl   recovery.Controller
	Pinger probe.Pinger
	// Now is injectable for tests; nil uses the real clock.
	Now func() time.Time
	// ProbeTicks and CheckTicks are injectable tick sources for tests. When
	// nil, Run creates real *time.Ticker channels from s.ProbeInterval and
	// s.CheckInterval respectively (and stops them on return).
	ProbeTicks <-chan time.Time
	CheckTicks <-chan time.Time
}

// Run blocks, driving handshake-age and (optional) probe triggers until ctx is
// cancelled. Recovery uses the resolve->reconnect episode from Phase 1.
func Run(ctx context.Context, d Deps, s *config.Settings, targets []string) {
	now := d.Now
	if now == nil {
		now = time.Now
	}
	opts := recovery.Options{
		Resolve: s.Resolve, ResolveRetries: s.ResolveRetries, ResolveBackoff: s.ResolveBackoff,
		Reconnect: s.Reconnect, ReconnectRetries: s.ReconnectRetries, ReconnectBackoff: s.ReconnectBackoff,
		OnAttempt: func(stage string, a int, err error) {
			if err != nil {
				log.Printf("bifrost: %s attempt %d: %v", stage, a, err)
			} else {
				log.Printf("bifrost: %s attempt %d", stage, a)
			}
		},
	}
	recover := func(reason string) {
		log.Printf("bifrost: %s — recovery", reason)
		if recovery.Recover(ctx, d.Ctrl, opts) {
			log.Print("bifrost: recovered")
		} else {
			log.Print("bifrost: recovery exhausted; will retry")
		}
	}

	probeOn := s.Probe && len(targets) > 0
	if s.Probe && len(targets) == 0 {
		log.Print("bifrost: probe enabled but no targets — probe inactive")
	}

	checkC := d.CheckTicks
	if checkC == nil {
		checkT := time.NewTicker(s.CheckInterval)
		defer checkT.Stop()
		checkC = checkT.C
	}

	var probeC <-chan time.Time
	if probeOn {
		probeC = d.ProbeTicks
		if probeC == nil {
			pt := time.NewTicker(s.ProbeInterval)
			defer pt.Stop()
			probeC = pt.C
		}
		log.Printf("bifrost: liveness probe on — %v", targets)
	}
	st := probe.NewState(s.ProbeFails)

	for {
		select {
		case <-ctx.Done():
			return
		case <-probeC:
			if d.Ctrl.NewestHandshake().IsZero() {
				continue // startup guard: only after first handshake
			}
			down := probe.AllDown(targets, s.ProbeTimeout, d.Pinger)
			if st.Round(down) {
				recover("probe all targets down")
			}
		case <-checkC:
			newest := d.Ctrl.NewestHandshake()
			if newest.IsZero() || now().Sub(newest) > s.StaleAfter {
				if !newest.IsZero() {
					recover("handshake stale")
				}
			}
		}
	}
}
