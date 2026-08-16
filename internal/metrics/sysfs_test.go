//go:build linux

package metrics

import (
	"os"
	"path/filepath"
	"testing"
)

func writeStat(t *testing.T, dir, name, value string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(value+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestReadIfaceStats(t *testing.T) {
	base := t.TempDir()
	statsDir := filepath.Join(base, "wg0", "statistics")
	if err := os.MkdirAll(statsDir, 0755); err != nil {
		t.Fatal(err)
	}

	want := IfaceStats{
		RxBytes:   1024,
		TxBytes:   2048,
		RxPackets: 10,
		TxPackets: 20,
		RxErrors:  1,
		TxErrors:  2,
		RxDropped: 3,
		TxDropped: 4,
	}

	writeStat(t, statsDir, "rx_bytes", "1024")
	writeStat(t, statsDir, "tx_bytes", "2048")
	writeStat(t, statsDir, "rx_packets", "10")
	writeStat(t, statsDir, "tx_packets", "20")
	writeStat(t, statsDir, "rx_errors", "1")
	writeStat(t, statsDir, "tx_errors", "2")
	writeStat(t, statsDir, "rx_dropped", "3")
	writeStat(t, statsDir, "tx_dropped", "4")

	got, err := readIfaceStatsFrom(base, "wg0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestReadIfaceStatsMissing(t *testing.T) {
	base := t.TempDir()

	_, err := readIfaceStatsFrom(base, "noexist")
	if err == nil {
		t.Fatal("expected error for non-existent interface, got nil")
	}
}

func TestReadIfaceStatsPartial(t *testing.T) {
	base := t.TempDir()
	statsDir := filepath.Join(base, "wg0", "statistics")
	if err := os.MkdirAll(statsDir, 0755); err != nil {
		t.Fatal(err)
	}

	writeStat(t, statsDir, "rx_bytes", "5000")
	writeStat(t, statsDir, "tx_bytes", "9000")

	got, err := readIfaceStatsFrom(base, "wg0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.RxBytes != 5000 {
		t.Errorf("RxBytes = %d, want 5000", got.RxBytes)
	}
	if got.TxBytes != 9000 {
		t.Errorf("TxBytes = %d, want 9000", got.TxBytes)
	}
	if got.RxPackets != 0 {
		t.Errorf("RxPackets = %d, want 0", got.RxPackets)
	}
	if got.TxPackets != 0 {
		t.Errorf("TxPackets = %d, want 0", got.TxPackets)
	}
	if got.RxErrors != 0 {
		t.Errorf("RxErrors = %d, want 0", got.RxErrors)
	}
	if got.TxErrors != 0 {
		t.Errorf("TxErrors = %d, want 0", got.TxErrors)
	}
	if got.RxDropped != 0 {
		t.Errorf("RxDropped = %d, want 0", got.RxDropped)
	}
	if got.TxDropped != 0 {
		t.Errorf("TxDropped = %d, want 0", got.TxDropped)
	}
}
