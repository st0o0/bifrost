//go:build !linux

package metrics

import "errors"

// IfaceStats holds kernel-level interface statistics from sysfs.
type IfaceStats struct {
	RxBytes   uint64
	TxBytes   uint64
	RxPackets uint64
	TxPackets uint64
	RxErrors  uint64
	TxErrors  uint64
	RxDropped uint64
	TxDropped uint64
}

// ReadIfaceStats is not available on non-Linux platforms.
func ReadIfaceStats(string) (IfaceStats, error) {
	return IfaceStats{}, errors.New("sysfs not available on this platform")
}
