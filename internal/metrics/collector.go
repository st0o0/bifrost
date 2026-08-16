package metrics

import (
	"runtime/debug"
	"strconv"
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
	deviceType string
	iface      string
	ifaceStats func(string) (IfaceStats, error)

	tunnelUp               *prometheus.Desc
	tunnelUpSince          *prometheus.Desc
	peerHandshakeAge       *prometheus.Desc
	peerReceiveBytesTotal  *prometheus.Desc
	peerTransmitBytesTotal *prometheus.Desc
	peerEndpointInfo       *prometheus.Desc
	peerAllowedIPsCount    *prometheus.Desc
	reconnectsTotal        *prometheus.Desc
	resolvesTotal          *prometheus.Desc
	endpointChangesTotal   *prometheus.Desc
	recoveryDuration       *prometheus.Desc
	recoveryInProgress     *prometheus.Desc
	resolveDuration        *prometheus.Desc
	probeRTT               *prometheus.Desc
	buildInfo              *prometheus.Desc
	deviceInfo             *prometheus.Desc
	ifaceRxBytes           *prometheus.Desc
	ifaceTxBytes           *prometheus.Desc
	ifaceRxPackets         *prometheus.Desc
	ifaceTxPackets         *prometheus.Desc
	ifaceRxErrors          *prometheus.Desc
	ifaceTxErrors          *prometheus.Desc
	ifaceRxDropped         *prometheus.Desc
	ifaceTxDropped         *prometheus.Desc
}

// CollectorOpts configures a new Collector.
type CollectorOpts struct {
	Stats      *Stats
	Device     DeviceFunc
	StaleAfter time.Duration
	ProbeOn    bool
	Now        func() time.Time
	DeviceType string
	Interface  string
	IfaceStats func(string) (IfaceStats, error)
}

// NewCollector creates a Collector. Now defaults to time.Now if nil.
func NewCollector(o CollectorOpts) *Collector {
	now := o.Now
	if now == nil {
		now = time.Now
	}
	ifaceStats := o.IfaceStats
	if ifaceStats == nil && o.Interface != "" {
		ifaceStats = ReadIfaceStats
	}
	c := &Collector{
		stats:      o.Stats,
		device:     o.Device,
		staleAfter: o.StaleAfter,
		probeOn:    o.ProbeOn,
		now:        now,
		deviceType: o.DeviceType,
		iface:      o.Interface,
		ifaceStats: ifaceStats,

		tunnelUp: prometheus.NewDesc("bifrost_tunnel_up",
			"Whether the WireGuard tunnel is up (1) or down (0).",
			nil, nil),
		tunnelUpSince: prometheus.NewDesc("bifrost_tunnel_up_since_seconds",
			"Unix timestamp of when the tunnel was last brought up successfully.",
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
		peerEndpointInfo: prometheus.NewDesc("bifrost_peer_endpoint_info",
			"Current endpoint address of the peer.",
			[]string{"peer", "endpoint"}, nil),
		peerAllowedIPsCount: prometheus.NewDesc("bifrost_peer_allowed_ips_count",
			"Number of allowed IP prefixes configured for the peer.",
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
		recoveryDuration: prometheus.NewDesc("bifrost_recovery_duration_seconds",
			"Wall-clock duration of the most recently completed recovery episode.",
			nil, nil),
		recoveryInProgress: prometheus.NewDesc("bifrost_recovery_in_progress",
			"Whether a recovery episode is currently executing (1) or not (0).",
			nil, nil),
		resolveDuration: prometheus.NewDesc("bifrost_resolve_duration_seconds",
			"Wall-clock duration of the most recent Resolve() call.",
			nil, nil),
		buildInfo: prometheus.NewDesc("bifrost_build_info",
			"Build information.",
			[]string{"version"}, nil),
		deviceInfo: prometheus.NewDesc("bifrost_device_info",
			"Device information.",
			[]string{"type"}, nil),
		ifaceRxBytes: prometheus.NewDesc("bifrost_interface_rx_bytes_total",
			"Cumulative bytes received on the WireGuard network interface.",
			nil, nil),
		ifaceTxBytes: prometheus.NewDesc("bifrost_interface_tx_bytes_total",
			"Cumulative bytes transmitted on the WireGuard network interface.",
			nil, nil),
		ifaceRxPackets: prometheus.NewDesc("bifrost_interface_rx_packets_total",
			"Cumulative packets received on the WireGuard network interface.",
			nil, nil),
		ifaceTxPackets: prometheus.NewDesc("bifrost_interface_tx_packets_total",
			"Cumulative packets transmitted on the WireGuard network interface.",
			nil, nil),
		ifaceRxErrors: prometheus.NewDesc("bifrost_interface_rx_errors_total",
			"Cumulative receive errors on the WireGuard network interface.",
			nil, nil),
		ifaceTxErrors: prometheus.NewDesc("bifrost_interface_tx_errors_total",
			"Cumulative transmit errors on the WireGuard network interface.",
			nil, nil),
		ifaceRxDropped: prometheus.NewDesc("bifrost_interface_rx_dropped_total",
			"Cumulative dropped received packets on the WireGuard network interface.",
			nil, nil),
		ifaceTxDropped: prometheus.NewDesc("bifrost_interface_tx_dropped_total",
			"Cumulative dropped transmitted packets on the WireGuard network interface.",
			nil, nil),
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
	ch <- c.tunnelUpSince
	ch <- c.peerHandshakeAge
	ch <- c.peerReceiveBytesTotal
	ch <- c.peerTransmitBytesTotal
	ch <- c.peerEndpointInfo
	ch <- c.peerAllowedIPsCount
	ch <- c.reconnectsTotal
	ch <- c.resolvesTotal
	ch <- c.endpointChangesTotal
	ch <- c.recoveryDuration
	ch <- c.recoveryInProgress
	ch <- c.resolveDuration
	ch <- c.buildInfo
	ch <- c.deviceInfo
	ch <- c.ifaceRxBytes
	ch <- c.ifaceTxBytes
	ch <- c.ifaceRxPackets
	ch <- c.ifaceTxPackets
	ch <- c.ifaceRxErrors
	ch <- c.ifaceTxErrors
	ch <- c.ifaceRxDropped
	ch <- c.ifaceTxDropped
	if c.probeRTT != nil {
		ch <- c.probeRTT
	}
}

// Collect implements prometheus.Collector.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	now := c.now()

	dev, err := c.device()
	tunnelUpVal := 0.0
	if err == nil && dev != nil {
		for i, p := range dev.Peers {
			label := strconv.Itoa(i)
			if !p.LastHandshakeTime.IsZero() {
				age := now.Sub(p.LastHandshakeTime).Seconds()
				ch <- prometheus.MustNewConstMetric(c.peerHandshakeAge, prometheus.GaugeValue, age, label)
				if now.Sub(p.LastHandshakeTime) <= c.staleAfter {
					tunnelUpVal = 1.0
				}
			}
			ch <- prometheus.MustNewConstMetric(c.peerReceiveBytesTotal, prometheus.CounterValue, float64(p.ReceiveBytes), label)
			ch <- prometheus.MustNewConstMetric(c.peerTransmitBytesTotal, prometheus.CounterValue, float64(p.TransmitBytes), label)
			if p.Endpoint != nil {
				ch <- prometheus.MustNewConstMetric(c.peerEndpointInfo, prometheus.GaugeValue, 1, label, p.Endpoint.String())
			}
			ch <- prometheus.MustNewConstMetric(c.peerAllowedIPsCount, prometheus.GaugeValue, float64(len(p.AllowedIPs)), label)
		}
	}
	ch <- prometheus.MustNewConstMetric(c.tunnelUp, prometheus.GaugeValue, tunnelUpVal)

	if upSince := c.stats.TunnelUpSince(); upSince != 0 {
		ch <- prometheus.MustNewConstMetric(c.tunnelUpSince, prometheus.GaugeValue, float64(upSince))
	} else {
		ch <- prometheus.MustNewConstMetric(c.tunnelUpSince, prometheus.GaugeValue, 0)
	}

	ch <- prometheus.MustNewConstMetric(c.reconnectsTotal, prometheus.CounterValue, float64(c.stats.Reconnects()))
	ch <- prometheus.MustNewConstMetric(c.resolvesTotal, prometheus.CounterValue, float64(c.stats.Resolves()))
	ch <- prometheus.MustNewConstMetric(c.endpointChangesTotal, prometheus.CounterValue, float64(c.stats.EndpointChanges()))

	ch <- prometheus.MustNewConstMetric(c.recoveryDuration, prometheus.GaugeValue, c.stats.LastRecoveryDuration())
	inProgress := 0.0
	if c.stats.RecoveryInProgress() {
		inProgress = 1.0
	}
	ch <- prometheus.MustNewConstMetric(c.recoveryInProgress, prometheus.GaugeValue, inProgress)
	ch <- prometheus.MustNewConstMetric(c.resolveDuration, prometheus.GaugeValue, c.stats.LastResolveDuration())

	if c.probeRTT != nil {
		ch <- prometheus.MustNewConstMetric(c.probeRTT, prometheus.GaugeValue, c.stats.ProbeRTT())
	}

	ch <- prometheus.MustNewConstMetric(c.buildInfo, prometheus.GaugeValue, 1, version())

	if c.deviceType != "" {
		ch <- prometheus.MustNewConstMetric(c.deviceInfo, prometheus.GaugeValue, 1, c.deviceType)
	}

	if c.ifaceStats != nil {
		if st, err := c.ifaceStats(c.iface); err == nil {
			ch <- prometheus.MustNewConstMetric(c.ifaceRxBytes, prometheus.CounterValue, float64(st.RxBytes))
			ch <- prometheus.MustNewConstMetric(c.ifaceTxBytes, prometheus.CounterValue, float64(st.TxBytes))
			ch <- prometheus.MustNewConstMetric(c.ifaceRxPackets, prometheus.CounterValue, float64(st.RxPackets))
			ch <- prometheus.MustNewConstMetric(c.ifaceTxPackets, prometheus.CounterValue, float64(st.TxPackets))
			ch <- prometheus.MustNewConstMetric(c.ifaceRxErrors, prometheus.CounterValue, float64(st.RxErrors))
			ch <- prometheus.MustNewConstMetric(c.ifaceTxErrors, prometheus.CounterValue, float64(st.TxErrors))
			ch <- prometheus.MustNewConstMetric(c.ifaceRxDropped, prometheus.CounterValue, float64(st.RxDropped))
			ch <- prometheus.MustNewConstMetric(c.ifaceTxDropped, prometheus.CounterValue, float64(st.TxDropped))
		}
	}
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
