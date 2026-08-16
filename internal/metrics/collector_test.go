package metrics

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func fakeKey(s string) wgtypes.Key {
	var k wgtypes.Key
	copy(k[:], s)
	return k
}

func TestCollectorEmitsAllMetrics(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	key := fakeKey("testpeer1234567890123456789012")

	stats := &Stats{}
	stats.IncReconnects()
	stats.IncResolves()
	stats.IncResolves()
	stats.IncEndpointChanges()
	stats.SetProbeRTT(0.012)
	stats.SetTunnelUpSince(1735689600)

	device := func() (*wgtypes.Device, error) {
		return &wgtypes.Device{
			Peers: []wgtypes.Peer{{
				PublicKey:         key,
				LastHandshakeTime: now.Add(-30 * time.Second),
				ReceiveBytes:      1048576,
				TransmitBytes:     524288,
				Endpoint:          &net.UDPAddr{IP: net.ParseIP("203.0.113.1"), Port: 51820},
				AllowedIPs: []net.IPNet{
					{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)},
				},
			}},
		}, nil
	}

	c := NewCollector(CollectorOpts{
		Stats:      stats,
		Device:     device,
		StaleAfter: 135 * time.Second,
		ProbeOn:    true,
		Now:        func() time.Time { return now },
		DeviceType: "kernel",
	})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(c)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}

	names := make(map[string]bool)
	for _, mf := range mfs {
		names[mf.GetName()] = true
	}

	expected := []string{
		"bifrost_tunnel_up",
		"bifrost_tunnel_up_since_seconds",
		"bifrost_peer_last_handshake_age_seconds",
		"bifrost_peer_receive_bytes_total",
		"bifrost_peer_transmit_bytes_total",
		"bifrost_peer_endpoint_info",
		"bifrost_peer_allowed_ips_count",
		"bifrost_reconnects_total",
		"bifrost_resolves_total",
		"bifrost_endpoint_changes_total",
		"bifrost_recovery_duration_seconds",
		"bifrost_recovery_in_progress",
		"bifrost_resolve_duration_seconds",
		"bifrost_probe_rtt_seconds",
		"bifrost_build_info",
		"bifrost_device_info",
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing metric %s", name)
		}
	}
}

func TestCollectorTunnelUpWhenFresh(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	key := fakeKey("testpeer1234567890123456789012")

	c := NewCollector(CollectorOpts{
		Stats: &Stats{},
		Device: func() (*wgtypes.Device, error) {
			return &wgtypes.Device{
				Peers: []wgtypes.Peer{{
					PublicKey:         key,
					LastHandshakeTime: now.Add(-30 * time.Second),
				}},
			}, nil
		},
		StaleAfter: 135 * time.Second,
		ProbeOn:    false,
		Now:        func() time.Time { return now },
	})

	want := `
# HELP bifrost_tunnel_up Whether the WireGuard tunnel is up (1) or down (0).
# TYPE bifrost_tunnel_up gauge
bifrost_tunnel_up 1
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want), "bifrost_tunnel_up"); err != nil {
		t.Error(err)
	}
}

func TestCollectorTunnelDownWhenStale(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	key := fakeKey("testpeer1234567890123456789012")

	c := NewCollector(CollectorOpts{
		Stats: &Stats{},
		Device: func() (*wgtypes.Device, error) {
			return &wgtypes.Device{
				Peers: []wgtypes.Peer{{
					PublicKey:         key,
					LastHandshakeTime: now.Add(-300 * time.Second),
				}},
			}, nil
		},
		StaleAfter: 135 * time.Second,
		ProbeOn:    false,
		Now:        func() time.Time { return now },
	})

	want := `
# HELP bifrost_tunnel_up Whether the WireGuard tunnel is up (1) or down (0).
# TYPE bifrost_tunnel_up gauge
bifrost_tunnel_up 0
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want), "bifrost_tunnel_up"); err != nil {
		t.Error(err)
	}
}

func TestCollectorProbeRTTOmittedWhenDisabled(t *testing.T) {
	c := NewCollector(CollectorOpts{
		Stats:      &Stats{},
		Device:     func() (*wgtypes.Device, error) { return &wgtypes.Device{}, nil },
		StaleAfter: 135 * time.Second,
		ProbeOn:    false,
		Now:        time.Now,
	})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(c)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, mf := range mfs {
		if mf.GetName() == "bifrost_probe_rtt_seconds" {
			t.Error("bifrost_probe_rtt_seconds should not be registered when probe is off")
		}
	}
}

func TestCollectorPeerLabelIsIndex(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	key0 := fakeKey("peer0___________________________")
	key1 := fakeKey("peer1___________________________")

	c := NewCollector(CollectorOpts{
		Stats: &Stats{},
		Device: func() (*wgtypes.Device, error) {
			return &wgtypes.Device{
				Peers: []wgtypes.Peer{
					{PublicKey: key0, ReceiveBytes: 100, TransmitBytes: 50},
					{PublicKey: key1, ReceiveBytes: 200, TransmitBytes: 100},
				},
			}, nil
		},
		StaleAfter: 135 * time.Second,
		Now:        func() time.Time { return now },
	})

	want := `
# HELP bifrost_peer_receive_bytes_total Total bytes received from the peer.
# TYPE bifrost_peer_receive_bytes_total counter
bifrost_peer_receive_bytes_total{peer="0"} 100
bifrost_peer_receive_bytes_total{peer="1"} 200
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want), "bifrost_peer_receive_bytes_total"); err != nil {
		t.Error(err)
	}
}

func TestCollectorDeviceInfo(t *testing.T) {
	c := NewCollector(CollectorOpts{
		Stats:      &Stats{},
		Device:     func() (*wgtypes.Device, error) { return &wgtypes.Device{}, nil },
		StaleAfter: 135 * time.Second,
		Now:        time.Now,
		DeviceType: "kernel",
	})

	want := `
# HELP bifrost_device_info Device information.
# TYPE bifrost_device_info gauge
bifrost_device_info{type="kernel"} 1
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want), "bifrost_device_info"); err != nil {
		t.Error(err)
	}
}

func TestCollectorDeviceInfoUserspace(t *testing.T) {
	c := NewCollector(CollectorOpts{
		Stats:      &Stats{},
		Device:     func() (*wgtypes.Device, error) { return &wgtypes.Device{}, nil },
		StaleAfter: 135 * time.Second,
		Now:        time.Now,
		DeviceType: "userspace",
	})

	want := `
# HELP bifrost_device_info Device information.
# TYPE bifrost_device_info gauge
bifrost_device_info{type="userspace"} 1
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want), "bifrost_device_info"); err != nil {
		t.Error(err)
	}
}

func TestCollectorPeerEndpointInfo(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	key := fakeKey("testpeer1234567890123456789012")

	c := NewCollector(CollectorOpts{
		Stats: &Stats{},
		Device: func() (*wgtypes.Device, error) {
			return &wgtypes.Device{
				Peers: []wgtypes.Peer{{
					PublicKey: key,
					Endpoint:  &net.UDPAddr{IP: net.ParseIP("203.0.113.1"), Port: 51820},
				}},
			}, nil
		},
		StaleAfter: 135 * time.Second,
		Now:        func() time.Time { return now },
	})

	want := `
# HELP bifrost_peer_endpoint_info Current endpoint address of the peer.
# TYPE bifrost_peer_endpoint_info gauge
bifrost_peer_endpoint_info{endpoint="203.0.113.1:51820",peer="0"} 1
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want), "bifrost_peer_endpoint_info"); err != nil {
		t.Error(err)
	}
}

func TestCollectorPeerEndpointInfoOmittedWhenNil(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	key := fakeKey("testpeer1234567890123456789012")

	c := NewCollector(CollectorOpts{
		Stats: &Stats{},
		Device: func() (*wgtypes.Device, error) {
			return &wgtypes.Device{
				Peers: []wgtypes.Peer{{PublicKey: key}},
			}, nil
		},
		StaleAfter: 135 * time.Second,
		Now:        func() time.Time { return now },
	})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(c)
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, mf := range mfs {
		if mf.GetName() == "bifrost_peer_endpoint_info" {
			t.Error("bifrost_peer_endpoint_info should not be emitted when endpoint is nil")
		}
	}
}

func TestCollectorAllowedIPsCount(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	key := fakeKey("testpeer1234567890123456789012")

	c := NewCollector(CollectorOpts{
		Stats: &Stats{},
		Device: func() (*wgtypes.Device, error) {
			return &wgtypes.Device{
				Peers: []wgtypes.Peer{{
					PublicKey: key,
					AllowedIPs: []net.IPNet{
						{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)},
						{IP: net.ParseIP("10.1.0.0"), Mask: net.CIDRMask(24, 32)},
						{IP: net.ParseIP("10.2.0.0"), Mask: net.CIDRMask(24, 32)},
					},
				}},
			}, nil
		},
		StaleAfter: 135 * time.Second,
		Now:        func() time.Time { return now },
	})

	want := `
# HELP bifrost_peer_allowed_ips_count Number of allowed IP prefixes configured for the peer.
# TYPE bifrost_peer_allowed_ips_count gauge
bifrost_peer_allowed_ips_count{peer="0"} 3
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want), "bifrost_peer_allowed_ips_count"); err != nil {
		t.Error(err)
	}
}

func TestCollectorRecoveryMetrics(t *testing.T) {
	stats := &Stats{}
	stats.SetLastRecoveryDuration(45.5)
	stats.SetRecoveryStart(1000000)
	stats.SetLastResolveDuration(0.25)

	c := NewCollector(CollectorOpts{
		Stats:      stats,
		Device:     func() (*wgtypes.Device, error) { return &wgtypes.Device{}, nil },
		StaleAfter: 135 * time.Second,
		Now:        time.Now,
	})

	want := `
# HELP bifrost_recovery_in_progress Whether a recovery episode is currently executing (1) or not (0).
# TYPE bifrost_recovery_in_progress gauge
bifrost_recovery_in_progress 1
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want), "bifrost_recovery_in_progress"); err != nil {
		t.Error(err)
	}

	want = `
# HELP bifrost_recovery_duration_seconds Wall-clock duration of the most recently completed recovery episode.
# TYPE bifrost_recovery_duration_seconds gauge
bifrost_recovery_duration_seconds 45.5
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want), "bifrost_recovery_duration_seconds"); err != nil {
		t.Error(err)
	}

	want = `
# HELP bifrost_resolve_duration_seconds Wall-clock duration of the most recent Resolve() call.
# TYPE bifrost_resolve_duration_seconds gauge
bifrost_resolve_duration_seconds 0.25
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want), "bifrost_resolve_duration_seconds"); err != nil {
		t.Error(err)
	}
}

func TestCollectorRecoveryNotInProgress(t *testing.T) {
	c := NewCollector(CollectorOpts{
		Stats:      &Stats{},
		Device:     func() (*wgtypes.Device, error) { return &wgtypes.Device{}, nil },
		StaleAfter: 135 * time.Second,
		Now:        time.Now,
	})

	want := `
# HELP bifrost_recovery_in_progress Whether a recovery episode is currently executing (1) or not (0).
# TYPE bifrost_recovery_in_progress gauge
bifrost_recovery_in_progress 0
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want), "bifrost_recovery_in_progress"); err != nil {
		t.Error(err)
	}
}

func TestCollectorTunnelUpSince(t *testing.T) {
	stats := &Stats{}
	stats.SetTunnelUpSince(1720000000)

	c := NewCollector(CollectorOpts{
		Stats:      stats,
		Device:     func() (*wgtypes.Device, error) { return &wgtypes.Device{}, nil },
		StaleAfter: 135 * time.Second,
		Now:        time.Now,
	})

	want := `
# HELP bifrost_tunnel_up_since_seconds Unix timestamp of when the tunnel was last brought up successfully.
# TYPE bifrost_tunnel_up_since_seconds gauge
bifrost_tunnel_up_since_seconds 1.72e+09
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want), "bifrost_tunnel_up_since_seconds"); err != nil {
		t.Error(err)
	}
}

func TestCollectorTunnelUpSinceZero(t *testing.T) {
	c := NewCollector(CollectorOpts{
		Stats:      &Stats{},
		Device:     func() (*wgtypes.Device, error) { return &wgtypes.Device{}, nil },
		StaleAfter: 135 * time.Second,
		Now:        time.Now,
	})

	want := `
# HELP bifrost_tunnel_up_since_seconds Unix timestamp of when the tunnel was last brought up successfully.
# TYPE bifrost_tunnel_up_since_seconds gauge
bifrost_tunnel_up_since_seconds 0
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want), "bifrost_tunnel_up_since_seconds"); err != nil {
		t.Error(err)
	}
}

func TestCollectorIfaceStats(t *testing.T) {
	fakeIfaceStats := func(string) (IfaceStats, error) {
		return IfaceStats{
			RxBytes:   2097152,
			TxBytes:   1048576,
			RxPackets: 15000,
			TxPackets: 12000,
			RxErrors:  3,
			TxErrors:  1,
			RxDropped: 5,
			TxDropped: 2,
		}, nil
	}

	c := NewCollector(CollectorOpts{
		Stats:      &Stats{},
		Device:     func() (*wgtypes.Device, error) { return &wgtypes.Device{}, nil },
		StaleAfter: 135 * time.Second,
		Now:        time.Now,
		Interface:  "wg0",
		IfaceStats: fakeIfaceStats,
	})

	want := `
# HELP bifrost_interface_rx_bytes_total Cumulative bytes received on the WireGuard network interface.
# TYPE bifrost_interface_rx_bytes_total counter
bifrost_interface_rx_bytes_total 2.097152e+06
# HELP bifrost_interface_tx_bytes_total Cumulative bytes transmitted on the WireGuard network interface.
# TYPE bifrost_interface_tx_bytes_total counter
bifrost_interface_tx_bytes_total 1.048576e+06
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want),
		"bifrost_interface_rx_bytes_total", "bifrost_interface_tx_bytes_total"); err != nil {
		t.Error(err)
	}

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(c)
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	ifaceMetrics := map[string]bool{
		"bifrost_interface_rx_packets_total": false,
		"bifrost_interface_tx_packets_total": false,
		"bifrost_interface_rx_errors_total":  false,
		"bifrost_interface_tx_errors_total":  false,
		"bifrost_interface_rx_dropped_total": false,
		"bifrost_interface_tx_dropped_total": false,
	}
	for _, mf := range mfs {
		if _, ok := ifaceMetrics[mf.GetName()]; ok {
			ifaceMetrics[mf.GetName()] = true
		}
	}
	for name, found := range ifaceMetrics {
		if !found {
			t.Errorf("missing interface metric %s", name)
		}
	}
}

func TestCollectorIfaceStatsOmittedOnError(t *testing.T) {
	fakeIfaceStats := func(string) (IfaceStats, error) {
		return IfaceStats{}, &net.OpError{Op: "read", Err: net.ErrClosed}
	}

	c := NewCollector(CollectorOpts{
		Stats:      &Stats{},
		Device:     func() (*wgtypes.Device, error) { return &wgtypes.Device{}, nil },
		StaleAfter: 135 * time.Second,
		Now:        time.Now,
		Interface:  "wg0",
		IfaceStats: fakeIfaceStats,
	})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(c)
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, mf := range mfs {
		if strings.HasPrefix(mf.GetName(), "bifrost_interface_") {
			t.Errorf("interface metric %s should not be emitted on error", mf.GetName())
		}
	}
}

func TestCollectorIfaceStatsOmittedWhenNoInterface(t *testing.T) {
	c := NewCollector(CollectorOpts{
		Stats:      &Stats{},
		Device:     func() (*wgtypes.Device, error) { return &wgtypes.Device{}, nil },
		StaleAfter: 135 * time.Second,
		Now:        time.Now,
	})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(c)
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, mf := range mfs {
		if strings.HasPrefix(mf.GetName(), "bifrost_interface_") {
			t.Errorf("interface metric %s should not be emitted without interface configured", mf.GetName())
		}
	}
}
