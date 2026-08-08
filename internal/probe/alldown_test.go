package probe

import (
	"testing"
	"time"
)

type fakePinger map[string]bool // target -> up

func (f fakePinger) Up(t string, _ time.Duration) bool { return f[t] }

func TestAllDown(t *testing.T) {
	r := AllDown([]string{"a", "b"}, time.Second, fakePinger{"a": true})
	if r.Down {
		t.Error("one up => not all down")
	}
	r = AllDown([]string{"a", "b"}, time.Second, fakePinger{})
	if !r.Down {
		t.Error("none up => all down")
	}
	r = AllDown(nil, time.Second, fakePinger{})
	if r.Down {
		t.Error("no targets => not all down")
	}
}

type fakeRTTPinger struct {
	results map[string]struct {
		up  bool
		rtt time.Duration
	}
}

func (f fakeRTTPinger) Up(target string, timeout time.Duration) bool {
	up, _ := f.Ping(target, timeout)
	return up
}

func (f fakeRTTPinger) Ping(target string, _ time.Duration) (bool, time.Duration) {
	r, ok := f.results[target]
	if !ok {
		return false, 0
	}
	return r.up, r.rtt
}

func TestAllDownWithRTT(t *testing.T) {
	p := fakeRTTPinger{results: map[string]struct {
		up  bool
		rtt time.Duration
	}{
		"a": {true, 10 * time.Millisecond},
		"b": {true, 5 * time.Millisecond},
		"c": {false, 0},
	}}
	r := AllDown([]string{"a", "b", "c"}, time.Second, p)
	if r.Down {
		t.Error("two up => not all down")
	}
	if r.BestRTT != 5*time.Millisecond {
		t.Errorf("BestRTT = %v, want 5ms", r.BestRTT)
	}
}

func TestAllDownAllFail(t *testing.T) {
	p := fakeRTTPinger{results: map[string]struct {
		up  bool
		rtt time.Duration
	}{
		"a": {false, 0},
	}}
	r := AllDown([]string{"a"}, time.Second, p)
	if !r.Down {
		t.Error("all fail => all down")
	}
	if r.BestRTT != 0 {
		t.Errorf("BestRTT = %v, want 0 when all down", r.BestRTT)
	}
}
