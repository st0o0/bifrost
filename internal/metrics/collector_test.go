package metrics

import (
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

	device := func() (*wgtypes.Device, error) {
		return &wgtypes.Device{
			Peers: []wgtypes.Peer{{
				PublicKey:         key,
				LastHandshakeTime: now.Add(-30 * time.Second),
				ReceiveBytes:      1048576,
				TransmitBytes:     524288,
			}},
		}, nil
	}

	c := NewCollector(CollectorOpts{
		Stats:      stats,
		Device:     device,
		StaleAfter: 135 * time.Second,
		ProbeOn:    true,
		Now:        func() time.Time { return now },
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
		"bifrost_peer_last_handshake_age_seconds",
		"bifrost_peer_receive_bytes_total",
		"bifrost_peer_transmit_bytes_total",
		"bifrost_reconnects_total",
		"bifrost_resolves_total",
		"bifrost_endpoint_changes_total",
		"bifrost_probe_rtt_seconds",
		"bifrost_build_info",
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

func TestPeerLabel(t *testing.T) {
	key := fakeKey("ABCDEFGHijklmnop12345678901234")
	label := peerLabel(key)
	if len(label) > 8 {
		t.Errorf("peer label too long: %q", label)
	}
}
