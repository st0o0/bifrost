package probe

import "time"

// Pinger checks whether a single target is reachable within timeout.
type Pinger interface {
	Up(target string, timeout time.Duration) bool
}

// ProbeResult holds the outcome of an AllDown round.
type ProbeResult struct {
	Down    bool
	BestRTT time.Duration
}

// RTTPinger extends Pinger with round-trip time measurement.
type RTTPinger interface {
	Pinger
	Ping(target string, timeout time.Duration) (up bool, rtt time.Duration)
}

// AllDown reports true only when there is at least one target and none of them
// are Up (a "round is down only if all targets fail"). If the Pinger also
// implements RTTPinger, the best (lowest) RTT across reachable targets is
// returned in the result.
func AllDown(targets []string, timeout time.Duration, p Pinger) ProbeResult {
	if len(targets) == 0 {
		return ProbeResult{}
	}
	rttP, hasRTT := p.(RTTPinger)
	var bestRTT time.Duration
	anyUp := false
	for _, t := range targets {
		if hasRTT {
			up, rtt := rttP.Ping(t, timeout)
			if up {
				anyUp = true
				if bestRTT == 0 || rtt < bestRTT {
					bestRTT = rtt
				}
			}
		} else {
			if p.Up(t, timeout) {
				anyUp = true
			}
		}
	}
	return ProbeResult{Down: !anyUp, BestRTT: bestRTT}
}
