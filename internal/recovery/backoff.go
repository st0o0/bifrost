package recovery

import "time"

const maxBackoff = 60 * time.Second

// Backoff returns the exponential backoff for a 1-based attempt with the given
// base, capped at 60s.
func Backoff(attempt int, base time.Duration) time.Duration {
	d := base
	for i := 1; i < attempt; i++ {
		d *= 2
		if d >= maxBackoff {
			return maxBackoff
		}
	}
	if d > maxBackoff {
		return maxBackoff
	}
	return d
}
