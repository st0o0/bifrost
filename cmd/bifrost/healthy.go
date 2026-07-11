package main

import "time"

// healthy reports whether the newest handshake is fresh enough. A zero newest
// time (no handshake yet) is unhealthy.
func healthy(newest, now time.Time, maxAge time.Duration) bool {
	if newest.IsZero() {
		return false
	}
	return now.Sub(newest) <= maxAge
}
