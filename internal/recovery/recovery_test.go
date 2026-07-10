package recovery

import (
	"context"
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	base := 5 * time.Second
	for _, c := range []struct {
		n    int
		want time.Duration
	}{
		{1, 5 * time.Second}, {2, 10 * time.Second}, {3, 20 * time.Second}, {5, 60 * time.Second},
	} {
		if got := Backoff(c.n, base); got != c.want {
			t.Errorf("Backoff(%d) = %s, want %s", c.n, got, c.want)
		}
	}
}

// fakeCtrl reports a new handshake after N resolve or reconnect calls.
type fakeCtrl struct {
	hs                     time.Time
	resolveN, reconnectN   int
	resolveAt, reconnectAt int
}

func (f *fakeCtrl) NewestHandshake() time.Time { return f.hs }
func (f *fakeCtrl) Resolve() error {
	f.resolveN++
	if f.resolveAt > 0 && f.resolveN >= f.resolveAt {
		f.hs = time.Now()
	}
	return nil
}
func (f *fakeCtrl) Reconnect() error {
	f.reconnectN++
	if f.reconnectAt > 0 && f.reconnectN >= f.reconnectAt {
		f.hs = time.Now()
	}
	return nil
}

func opts() Options {
	return Options{
		Resolve: true, ResolveRetries: 3, ResolveBackoff: time.Second,
		Reconnect: true, ReconnectRetries: 3, ReconnectBackoff: time.Second,
		Sleep: func(context.Context, time.Duration) {}, // no real waiting
	}
}

func TestRecoverViaResolve(t *testing.T) {
	f := &fakeCtrl{resolveAt: 2}
	if !Recover(context.Background(), f, opts()) {
		t.Fatal("expected recovery via resolve")
	}
	if f.reconnectN != 0 {
		t.Errorf("reconnect ran %d times, expected 0", f.reconnectN)
	}
}

func TestRecoverViaReconnect(t *testing.T) {
	f := &fakeCtrl{reconnectAt: 2}
	if !Recover(context.Background(), f, opts()) {
		t.Fatal("expected recovery via reconnect")
	}
	if f.resolveN != 3 {
		t.Errorf("resolve ran %d times, expected 3 (exhausted first)", f.resolveN)
	}
}

func TestRecoverExhausted(t *testing.T) {
	f := &fakeCtrl{} // never recovers
	if Recover(context.Background(), f, opts()) {
		t.Fatal("expected failure")
	}
}

func TestRecoverContextCancelled(t *testing.T) {
	f := &fakeCtrl{} // never recovers
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if Recover(ctx, f, opts()) {
		t.Fatal("expected failure")
	}
	if f.resolveN+f.reconnectN > 1 {
		t.Errorf("attempts = %d, want <= 1 (abort almost immediately)", f.resolveN+f.reconnectN)
	}
}
