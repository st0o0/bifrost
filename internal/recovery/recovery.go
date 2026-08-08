package recovery

import (
	"context"
	"time"
)

// Controller is the tunnel control surface the recovery logic needs.
type Controller interface {
	NewestHandshake() time.Time // zero if none yet
	Resolve() error             // re-resolve endpoints in place (gentle)
	Reconnect() error           // rebuild the interface (down + up)
}

// Options configures one recovery episode.
type Options struct {
	Resolve          bool
	ResolveRetries   int
	ResolveBackoff   time.Duration
	Reconnect        bool
	ReconnectRetries int
	ReconnectBackoff time.Duration
	// Sleep is injectable for tests; nil uses a context-aware timer.
	Sleep func(context.Context, time.Duration)
	// OnAttempt, if set, is called after each attempt with the stage
	// ("resolve"/"reconnect"), the 1-based attempt number, and the action's
	// error (nil on success). For logging; must not block.
	OnAttempt func(stage string, attempt int, err error)
	// OnResolve is called after each resolve attempt (regardless of outcome).
	OnResolve func()
	// OnReconnect is called after each reconnect attempt (regardless of outcome).
	OnReconnect func()
}

// Recover runs one escalating recovery episode: resolve retries, then reconnect
// retries. Returns true when a handshake strictly newer than the baseline at
// entry is observed. Returns false immediately if ctx is cancelled.
//
// Note: NewestHandshake is device-wide (newest across peers). bifrost is a
// single-active-peer client supervisor; with multiple peers a healthy peer's
// routine handshake can satisfy the "recovered" test coarsely.
func Recover(ctx context.Context, c Controller, o Options) bool {
	baseline := c.NewestHandshake()
	sleep := o.Sleep
	if sleep == nil {
		sleep = ctxSleep
	}
	recovered := func() bool { return c.NewestHandshake().After(baseline) }

	run := func(stage string, retries int, base time.Duration, act func() error, onCall func()) bool {
		for a := 1; a <= retries; a++ {
			if ctx.Err() != nil {
				return false
			}
			err := act()
			if o.OnAttempt != nil {
				o.OnAttempt(stage, a, err)
			}
			if onCall != nil {
				onCall()
			}
			sleep(ctx, Backoff(a, base))
			if recovered() {
				return true
			}
		}
		return false
	}

	if o.Resolve && run("resolve", o.ResolveRetries, o.ResolveBackoff, c.Resolve, o.OnResolve) {
		return true
	}
	if o.Reconnect && run("reconnect", o.ReconnectRetries, o.ReconnectBackoff, c.Reconnect, o.OnReconnect) {
		return true
	}
	return false
}

func ctxSleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
