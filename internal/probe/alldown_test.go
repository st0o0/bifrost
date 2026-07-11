package probe

import (
	"testing"
	"time"
)

type fakePinger map[string]bool // target -> up

func (f fakePinger) Up(t string, _ time.Duration) bool { return f[t] }

func TestAllDown(t *testing.T) {
	if AllDown([]string{"a", "b"}, time.Second, fakePinger{"a": true}) {
		t.Error("one up => not all down")
	}
	if !AllDown([]string{"a", "b"}, time.Second, fakePinger{}) {
		t.Error("none up => all down")
	}
	if AllDown(nil, time.Second, fakePinger{}) {
		t.Error("no targets => not all down")
	}
}
