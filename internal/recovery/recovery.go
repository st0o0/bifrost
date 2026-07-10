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
}

// Recover runs one escalating recovery episode: resolve retries, then reconnect
// retries. Returns true when a handshake strictly newer than the baseline at
// entry is observed.
func Recover(ctx context.Context, c Controller, o Options) bool {
	baseline := c.NewestHandshake()
	sleep := o.Sleep
	if sleep == nil {
		sleep = ctxSleep
	}
	recovered := func() bool { return c.NewestHandshake().After(baseline) }

	if o.Resolve {
		for a := 1; a <= o.ResolveRetries; a++ {
			_ = c.Resolve()
			sleep(ctx, Backoff(a, o.ResolveBackoff))
			if recovered() {
				return true
			}
		}
	}
	if o.Reconnect {
		for a := 1; a <= o.ReconnectRetries; a++ {
			_ = c.Reconnect()
			sleep(ctx, Backoff(a, o.ReconnectBackoff))
			if recovered() {
				return true
			}
		}
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
