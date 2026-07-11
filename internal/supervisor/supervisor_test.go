package supervisor

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/st0o0/bifrost/internal/config"
)

// fakeCtrl is a recovery.Controller test double. Resolve/Reconnect record how
// often they were called and immediately advance the handshake clock so
// recovery.Recover reports success on the first attempt (keeping tests fast
// without needing to fake recovery's internal sleep). notify fires (non-
// blocking) on every Resolve/Reconnect call so tests can synchronize without
// sleeping.
type fakeCtrl struct {
	mu         sync.Mutex
	hs         time.Time
	resolveN   int
	reconnectN int
	notify     chan struct{}
}

func newFakeCtrl(hs time.Time) *fakeCtrl {
	return &fakeCtrl{hs: hs, notify: make(chan struct{}, 16)}
}

func (f *fakeCtrl) NewestHandshake() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.hs
}

func (f *fakeCtrl) Resolve() error {
	f.mu.Lock()
	f.resolveN++
	f.hs = time.Now()
	f.mu.Unlock()
	f.ping()
	return nil
}

func (f *fakeCtrl) Reconnect() error {
	f.mu.Lock()
	f.reconnectN++
	f.hs = time.Now()
	f.mu.Unlock()
	f.ping()
	return nil
}

func (f *fakeCtrl) ping() {
	select {
	case f.notify <- struct{}{}:
	default:
	}
}

func (f *fakeCtrl) counts() (resolve, reconnect int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.resolveN, f.reconnectN
}

// awaitRecovery waits (with a timeout) for a Resolve/Reconnect call.
func awaitRecovery(t *testing.T, f *fakeCtrl) {
	t.Helper()
	select {
	case <-f.notify:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for recovery (Resolve/Reconnect) to be invoked")
	}
}

// assertNoRecovery makes sure no Resolve/Reconnect call happens within a
// short grace period.
func assertNoRecovery(t *testing.T, f *fakeCtrl) {
	t.Helper()
	select {
	case <-f.notify:
		t.Fatal("recovery was invoked but should not have been")
	case <-time.After(150 * time.Millisecond):
	}
}

// fakePinger reports every target as up or down according to a fixed value.
type fakePinger struct{ up bool }

func (p fakePinger) Up(string, time.Duration) bool { return p.up }

func fastOpts() (resolveRetries, reconnectRetries int, resolveBackoff, reconnectBackoff time.Duration) {
	return 1, 1, 0, 0
}

// baseSettings returns settings with fast/no-op backoffs so recovery episodes
// resolve (pun intended) quickly in tests.
func baseSettings() *config.Settings {
	rr, cr, rb, cb := fastOpts()
	return &config.Settings{
		CheckInterval:    time.Hour, // unused: tests drive via injected channels
		StaleAfter:       time.Minute,
		Resolve:          true,
		ResolveRetries:   rr,
		ResolveBackoff:   rb,
		Reconnect:        true,
		ReconnectRetries: cr,
		ReconnectBackoff: cb,
		Probe:            false,
		ProbeInterval:    time.Hour,
		ProbeFails:       3,
		ProbeTimeout:     time.Second,
	}
}

// runInBackground starts Run in a goroutine and returns a channel closed
// once Run returns (i.e. after ctx is cancelled).
func runInBackground(ctx context.Context, d Deps, s *config.Settings, targets []string) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		Run(ctx, d, s, targets)
		close(done)
	}()
	return done
}

func waitDone(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after ctx cancellation")
	}
}

// TestProbeTriggersRecovery drives PROBE_FAILS consecutive down rounds via
// the probe ticker and asserts recovery ran.
func TestProbeTriggersRecovery(t *testing.T) {
	s := baseSettings()
	s.Probe = true
	s.ProbeFails = 3
	targets := []string{"10.0.0.1"}

	ctrl := newFakeCtrl(time.Now().Add(-time.Hour)) // non-zero, past: startup guard passes
	probeTicks := make(chan time.Time)
	checkTicks := make(chan time.Time) // never fed; keeps the stale trigger silent

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	d := Deps{
		Ctrl:       ctrl,
		Pinger:     fakePinger{up: false}, // every target down
		ProbeTicks: probeTicks,
		CheckTicks: checkTicks,
	}
	done := runInBackground(ctx, d, s, targets)

	for i := 0; i < s.ProbeFails; i++ {
		probeTicks <- time.Now()
	}
	awaitRecovery(t, ctrl)

	cancel()
	waitDone(t, done)

	resolveN, reconnectN := ctrl.counts()
	if resolveN == 0 {
		t.Errorf("resolveN = 0, want recovery to have attempted resolve")
	}
	if reconnectN != 0 {
		t.Errorf("reconnectN = %d, want 0 (resolve should have recovered first)", reconnectN)
	}
}

// TestStaleHandshakeTriggersRecovery has the probe off entirely; a check
// tick with a handshake older than StaleAfter must trigger recovery.
func TestStaleHandshakeTriggersRecovery(t *testing.T) {
	s := baseSettings()
	s.Probe = false // probe trigger fully disabled
	s.StaleAfter = time.Minute

	fixedNow := time.Now()
	ctrl := newFakeCtrl(fixedNow.Add(-time.Hour)) // well past StaleAfter
	checkTicks := make(chan time.Time)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	d := Deps{
		Ctrl:       ctrl,
		Pinger:     fakePinger{up: true},
		Now:        func() time.Time { return fixedNow },
		CheckTicks: checkTicks,
	}
	done := runInBackground(ctx, d, s, nil)

	checkTicks <- time.Now()
	awaitRecovery(t, ctrl)

	cancel()
	waitDone(t, done)

	resolveN, reconnectN := ctrl.counts()
	if resolveN+reconnectN == 0 {
		t.Errorf("no recovery attempts observed, want at least one")
	}
}

// TestStartupGuardBlocksProbe ensures a zero NewestHandshake (no handshake
// yet observed) suppresses the probe trigger, even across many down rounds.
func TestStartupGuardBlocksProbe(t *testing.T) {
	s := baseSettings()
	s.Probe = true
	s.ProbeFails = 2
	targets := []string{"10.0.0.1"}

	ctrl := newFakeCtrl(time.Time{}) // zero: no handshake observed yet
	probeTicks := make(chan time.Time)
	checkTicks := make(chan time.Time)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	d := Deps{
		Ctrl:       ctrl,
		Pinger:     fakePinger{up: false},
		ProbeTicks: probeTicks,
		CheckTicks: checkTicks,
	}
	done := runInBackground(ctx, d, s, targets)

	// Feed well more ticks than ProbeFails; without the guard this would trigger.
	for i := 0; i < s.ProbeFails*3; i++ {
		probeTicks <- time.Now()
	}
	assertNoRecovery(t, ctrl)

	cancel()
	waitDone(t, done)

	resolveN, reconnectN := ctrl.counts()
	if resolveN+reconnectN != 0 {
		t.Errorf("recovery attempts = %d, want 0 (startup guard should block)", resolveN+reconnectN)
	}
}

// TestEmptyTargetsDisablesProbe ensures that with Probe on but no targets,
// the probe ticker is never wired up (nothing reads from d.ProbeTicks).
func TestEmptyTargetsDisablesProbe(t *testing.T) {
	s := baseSettings()
	s.Probe = true
	s.ProbeFails = 1

	ctrl := newFakeCtrl(time.Now())
	probeTicks := make(chan time.Time)
	checkTicks := make(chan time.Time)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	d := Deps{
		Ctrl:       ctrl,
		Pinger:     fakePinger{up: false},
		ProbeTicks: probeTicks,
		CheckTicks: checkTicks,
	}
	done := runInBackground(ctx, d, s, nil) // no targets

	select {
	case probeTicks <- time.Now():
		t.Fatal("probe ticker was consumed even though there are no targets")
	case <-time.After(150 * time.Millisecond):
		// expected: nobody is listening on probeTicks
	}

	cancel()
	waitDone(t, done)
}

// TestReconnectDoesNotGoDormant simulates the DDNS-recovery scenario from
// I-1: the handshake starts non-zero (latching everHandshaked), then a
// reconnect leaves it transiently zero (e.g. the rebuilt interface hasn't
// re-handshaked yet because DDNS hasn't propagated). A subsequent check tick
// must still trigger recovery — the loop must not go permanently dormant
// just because NewestHandshake() is momentarily zero.
func TestReconnectDoesNotGoDormant(t *testing.T) {
	s := baseSettings()
	s.Probe = false
	s.StaleAfter = time.Minute

	ctrl := newFakeCtrl(time.Now()) // non-zero: latches everHandshaked
	checkTicks := make(chan time.Time)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	d := Deps{
		Ctrl:       ctrl,
		Pinger:     fakePinger{up: true},
		CheckTicks: checkTicks,
	}
	done := runInBackground(ctx, d, s, nil)

	// Simulate a post-reconnect device that hasn't re-handshaked yet: the
	// handshake goes back to zero, as it would right after Reconnect()
	// rebuilds the interface.
	ctrl.mu.Lock()
	ctrl.hs = time.Time{}
	ctrl.mu.Unlock()

	checkTicks <- time.Now()
	awaitRecovery(t, ctrl)

	cancel()
	waitDone(t, done)

	resolveN, reconnectN := ctrl.counts()
	if resolveN+reconnectN == 0 {
		t.Errorf("no recovery attempts observed after zero handshake post-reconnect, want at least one (loop must not go dormant)")
	}
}

// TestProbeReArmsAfterTransientZeroHandshake ensures that once the
// everHandshaked latch is set, a later period where NewestHandshake() reads
// zero (e.g. right after a reconnect) does not suppress the probe trigger:
// the probe must still recover on consecutive down rounds.
func TestProbeReArmsAfterTransientZeroHandshake(t *testing.T) {
	s := baseSettings()
	s.Probe = true
	s.ProbeFails = 2
	targets := []string{"10.0.0.1"}

	ctrl := newFakeCtrl(time.Now().Add(-time.Hour)) // non-zero: latches everHandshaked
	probeTicks := make(chan time.Time)
	checkTicks := make(chan time.Time) // never fed; keeps the stale trigger silent

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	d := Deps{
		Ctrl:       ctrl,
		Pinger:     fakePinger{up: false}, // every target down
		ProbeTicks: probeTicks,
		CheckTicks: checkTicks,
	}
	done := runInBackground(ctx, d, s, targets)

	// First probe tick observes the non-zero handshake, setting the latch.
	probeTicks <- time.Now()

	// Now simulate the transient zero window right after a reconnect.
	ctrl.mu.Lock()
	ctrl.hs = time.Time{}
	ctrl.mu.Unlock()

	// Feed the remaining down rounds; the probe must not be re-suppressed by
	// the transient zero handshake now that the latch is set.
	for i := 1; i < s.ProbeFails; i++ {
		probeTicks <- time.Now()
	}
	awaitRecovery(t, ctrl)

	cancel()
	waitDone(t, done)

	resolveN, reconnectN := ctrl.counts()
	if resolveN+reconnectN == 0 {
		t.Errorf("no recovery attempts observed, want probe to re-arm despite transient zero handshake")
	}
}

// TestCtxCancelStopsLoop verifies Run returns promptly once ctx is cancelled,
// with no ticks delivered at all.
func TestCtxCancelStopsLoop(t *testing.T) {
	s := baseSettings()
	ctrl := newFakeCtrl(time.Now())
	ctx, cancel := context.WithCancel(context.Background())
	d := Deps{
		Ctrl:       ctrl,
		Pinger:     fakePinger{up: true},
		CheckTicks: make(chan time.Time),
	}
	done := runInBackground(ctx, d, s, nil)
	cancel()
	waitDone(t, done)
}
