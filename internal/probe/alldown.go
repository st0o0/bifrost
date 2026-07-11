package probe

import "time"

// Pinger checks whether a single target is reachable within timeout.
type Pinger interface {
	Up(target string, timeout time.Duration) bool
}

// AllDown reports true only when there is at least one target and none of them
// are Up (a "round is down only if all targets fail").
func AllDown(targets []string, timeout time.Duration, p Pinger) bool {
	if len(targets) == 0 {
		return false
	}
	for _, t := range targets {
		if p.Up(t, timeout) {
			return false
		}
	}
	return true
}
