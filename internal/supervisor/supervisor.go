// Package supervisor drives recovery from two independent triggers: the
// liveness probe (all targets down for N consecutive rounds) and
// handshake-age staleness. It depends only on the recovery.Controller and
// probe.Pinger interfaces plus Phase 1 packages, so it is cross-platform and
// unit-testable without a real WireGuard interface.
package supervisor

import (
	"context"
	"log/slog"
	"time"

	"github.com/st0o0/bifrost/internal/config"
	"github.com/st0o0/bifrost/internal/metrics"
	"github.com/st0o0/bifrost/internal/probe"
	"github.com/st0o0/bifrost/internal/recovery"
)

// Deps bundles the supervisor's collaborators and injectable timing sources.
type Deps struct {
	Ctrl   recovery.Controller
	Pinger probe.Pinger
	// Stats receives metric counter increments; nil disables metrics tracking.
	Stats *metrics.Stats
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
				slog.Warn("recovery attempt failed", "stage", stage, "attempt", a, "error", err)
			} else {
				slog.Info("recovery attempt", "stage", stage, "attempt", a)
			}
		},
	}
	ctrl := d.Ctrl
	if d.Stats != nil {
		opts.OnResolve = d.Stats.IncResolves
		opts.OnReconnect = d.Stats.IncReconnects
		ctrl = &timedCtrl{Controller: d.Ctrl, stats: d.Stats}
	}
	recoverNow := func(reason string) {
		slog.Warn("starting recovery", "reason", reason)
		if d.Stats != nil {
			d.Stats.SetRecoveryStart(now().UnixNano())
		}
		start := now()
		ok := recovery.Recover(ctx, ctrl, opts)
		if d.Stats != nil {
			d.Stats.SetLastRecoveryDuration(time.Since(start).Seconds())
			d.Stats.SetRecoveryStart(0)
			if ok {
				d.Stats.SetTunnelUpSince(now().Unix())
			}
		}
		if ok {
			slog.Info("recovered")
		} else {
			slog.Warn("recovery exhausted; will retry")
		}
	}

	probeOn := s.Probe && len(targets) > 0
	if s.Probe && len(targets) == 0 {
		slog.Info("probe enabled but no targets")
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
		slog.Info("liveness probe on", "targets", targets)
	}
	st := probe.NewState(s.ProbeFails)

	// everHandshaked latches true the first time NewestHandshake() is
	// observed to be non-zero. Unlike the instantaneous handshake check, it
	// never resets — so a reconnect that leaves the handshake transiently
	// zero (e.g. DDNS not yet propagated) does not permanently silence
	// either trigger.
	var everHandshaked bool

	for {
		select {
		case <-ctx.Done():
			return
		case <-probeC:
			hs := d.Ctrl.NewestHandshake()
			if !hs.IsZero() {
				everHandshaked = true
			}
			if !everHandshaked {
				continue // startup guard: only after the first-ever handshake
			}
			result := probe.AllDown(targets, s.ProbeTimeout, d.Pinger)
			if d.Stats != nil && result.BestRTT > 0 {
				d.Stats.SetProbeRTT(result.BestRTT.Seconds())
			}
			if st.Round(result.Down) {
				recoverNow("probe all targets down")
			}
		case <-checkC:
			newest := d.Ctrl.NewestHandshake()
			if !newest.IsZero() {
				everHandshaked = true
			}
			if newest.IsZero() || now().Sub(newest) > s.StaleAfter {
				recoverNow("handshake stale")
			}
		}
	}
}

type timedCtrl struct {
	recovery.Controller
	stats *metrics.Stats
}

func (t *timedCtrl) Resolve() error {
	start := time.Now()
	err := t.Controller.Resolve()
	t.stats.SetLastResolveDuration(time.Since(start).Seconds())
	return err
}
