package metrics

import (
	"runtime/debug"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// DeviceFunc returns the current WireGuard device state.
type DeviceFunc func() (*wgtypes.Device, error)

// Collector implements prometheus.Collector, reading live WireGuard state on
// each scrape and combining it with process-lifetime counters from Stats.
type Collector struct {
	stats      *Stats
	device     DeviceFunc
	staleAfter time.Duration
	probeOn    bool
	now        func() time.Time

	tunnelUp              *prometheus.Desc
	peerHandshakeAge      *prometheus.Desc
	peerReceiveBytesTotal *prometheus.Desc
	peerTransmitBytesTotal *prometheus.Desc
	reconnectsTotal       *prometheus.Desc
	resolvesTotal         *prometheus.Desc
	endpointChangesTotal  *prometheus.Desc
	probeRTT              *prometheus.Desc
	buildInfo             *prometheus.Desc
}

// CollectorOpts configures a new Collector.
type CollectorOpts struct {
	Stats      *Stats
	Device     DeviceFunc
	StaleAfter time.Duration
	ProbeOn    bool
	Now        func() time.Time
}

// NewCollector creates a Collector. Now defaults to time.Now if nil.
func NewCollector(o CollectorOpts) *Collector {
	now := o.Now
	if now == nil {
		now = time.Now
	}
	c := &Collector{
		stats:      o.Stats,
		device:     o.Device,
		staleAfter: o.StaleAfter,
		probeOn:    o.ProbeOn,
		now:        now,

		tunnelUp: prometheus.NewDesc("bifrost_tunnel_up",
			"Whether the WireGuard tunnel is up (1) or down (0).",
			nil, nil),
		peerHandshakeAge: prometheus.NewDesc("bifrost_peer_last_handshake_age_seconds",
			"Seconds since the peer's last WireGuard handshake.",
			[]string{"peer"}, nil),
		peerReceiveBytesTotal: prometheus.NewDesc("bifrost_peer_receive_bytes_total",
			"Total bytes received from the peer.",
			[]string{"peer"}, nil),
		peerTransmitBytesTotal: prometheus.NewDesc("bifrost_peer_transmit_bytes_total",
			"Total bytes transmitted to the peer.",
			[]string{"peer"}, nil),
		reconnectsTotal: prometheus.NewDesc("bifrost_reconnects_total",
			"Total number of reconnect attempts.",
			nil, nil),
		resolvesTotal: prometheus.NewDesc("bifrost_resolves_total",
			"Total number of DDNS re-resolution attempts.",
			nil, nil),
		endpointChangesTotal: prometheus.NewDesc("bifrost_endpoint_changes_total",
			"Total number of endpoint IP changes detected during DDNS re-resolution.",
			nil, nil),
		buildInfo: prometheus.NewDesc("bifrost_build_info",
			"Build information.",
			[]string{"version"}, nil),
	}
	if o.ProbeOn {
		c.probeRTT = prometheus.NewDesc("bifrost_probe_rtt_seconds",
			"Round-trip time of the most recent liveness probe in seconds.",
			nil, nil)
	}
	return c
}

// Describe implements prometheus.Collector.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.tunnelUp
	ch <- c.peerHandshakeAge
	ch <- c.peerReceiveBytesTotal
	ch <- c.peerTransmitBytesTotal
	ch <- c.reconnectsTotal
	ch <- c.resolvesTotal
	ch <- c.endpointChangesTotal
	ch <- c.buildInfo
	if c.probeRTT != nil {
		ch <- c.probeRTT
	}
}

func peerLabel(key wgtypes.Key) string {
	s := key.String()
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

// Collect implements prometheus.Collector.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	now := c.now()

	dev, err := c.device()
	tunnelUpVal := 0.0
	if err == nil && dev != nil {
		for _, p := range dev.Peers {
			label := peerLabel(p.PublicKey)
			if !p.LastHandshakeTime.IsZero() {
				age := now.Sub(p.LastHandshakeTime).Seconds()
				ch <- prometheus.MustNewConstMetric(c.peerHandshakeAge, prometheus.GaugeValue, age, label)
				if now.Sub(p.LastHandshakeTime) <= c.staleAfter {
					tunnelUpVal = 1.0
				}
			}
			ch <- prometheus.MustNewConstMetric(c.peerReceiveBytesTotal, prometheus.CounterValue, float64(p.ReceiveBytes), label)
			ch <- prometheus.MustNewConstMetric(c.peerTransmitBytesTotal, prometheus.CounterValue, float64(p.TransmitBytes), label)
		}
	}
	ch <- prometheus.MustNewConstMetric(c.tunnelUp, prometheus.GaugeValue, tunnelUpVal)

	ch <- prometheus.MustNewConstMetric(c.reconnectsTotal, prometheus.CounterValue, float64(c.stats.Reconnects()))
	ch <- prometheus.MustNewConstMetric(c.resolvesTotal, prometheus.CounterValue, float64(c.stats.Resolves()))
	ch <- prometheus.MustNewConstMetric(c.endpointChangesTotal, prometheus.CounterValue, float64(c.stats.EndpointChanges()))

	if c.probeRTT != nil {
		ch <- prometheus.MustNewConstMetric(c.probeRTT, prometheus.GaugeValue, c.stats.ProbeRTT())
	}

	ch <- prometheus.MustNewConstMetric(c.buildInfo, prometheus.GaugeValue, 1, version())
}

func version() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return "dev"
}
