package probe

// State counts consecutive "all targets down" rounds and reports when recovery
// should trigger (>= threshold consecutive down rounds). The counter resets on
// any up round and after a trigger.
type State struct {
	fails     int
	threshold int
}

func NewState(threshold int) *State { return &State{threshold: threshold} }

// Round records one round's result (down = every target failed) and returns
// true if recovery should be triggered now.
func (s *State) Round(down bool) bool {
	if !down {
		s.fails = 0
		return false
	}
	s.fails++
	if s.fails >= s.threshold {
		s.fails = 0
		return true
	}
	return false
}
