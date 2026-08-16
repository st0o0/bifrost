package metrics

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

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

// ReadIfaceStats reads interface statistics from /sys/class/net/<iface>/statistics/.
func ReadIfaceStats(iface string) (IfaceStats, error) {
	return readIfaceStatsFrom("/sys/class/net", iface)
}

func readIfaceStatsFrom(base, iface string) (IfaceStats, error) {
	dir := filepath.Join(base, iface, "statistics")
	if _, err := os.Stat(dir); err != nil {
		return IfaceStats{}, fmt.Errorf("sysfs stats dir: %w", err)
	}
	var s IfaceStats
	s.RxBytes, _ = readSysfsUint64(filepath.Join(dir, "rx_bytes"))
	s.TxBytes, _ = readSysfsUint64(filepath.Join(dir, "tx_bytes"))
	s.RxPackets, _ = readSysfsUint64(filepath.Join(dir, "rx_packets"))
	s.TxPackets, _ = readSysfsUint64(filepath.Join(dir, "tx_packets"))
	s.RxErrors, _ = readSysfsUint64(filepath.Join(dir, "rx_errors"))
	s.TxErrors, _ = readSysfsUint64(filepath.Join(dir, "tx_errors"))
	s.RxDropped, _ = readSysfsUint64(filepath.Join(dir, "rx_dropped"))
	s.TxDropped, _ = readSysfsUint64(filepath.Join(dir, "tx_dropped"))
	return s, nil
}

func readSysfsUint64(path string) (uint64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
}
