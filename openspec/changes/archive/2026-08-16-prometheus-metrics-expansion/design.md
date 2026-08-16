## Context

bifrost already has a working Prometheus metrics endpoint (`internal/metrics/`) with a custom `prometheus.Collector` that reads live `wgtypes.Device` state on each scrape. The `Stats` struct holds process-lifetime atomic counters incremented by the supervisor. This change expands both the Collector and Stats to surface five new metric groups: device metadata, peer metadata, kernel interface stats, recovery timing, and tunnel uptime.

The existing architecture — Collector reads live device state + Stats counters, supervisor increments Stats — extends naturally to all five groups.

## Goals / Non-Goals

**Goals:**
- Surface all operationally useful WireGuard data via Prometheus without requiring `wg show` or shell access.
- Give operators recovery observability (duration, in-progress, resolve latency) for alerting.
- Provide kernel-level interface statistics that WireGuard peer counters don't capture (drops, errors).
- Track continuous tunnel uptime for SLA reporting.

**Non-Goals:**
- Push-based metrics, OTLP, or any non-pull model.
- Exposing sensitive material (private keys, preshared keys) in any form.
- Per-peer endpoint history or cardinality management — one info gauge per peer, label changes on DDNS flip.
- Graceful shutdown of the metrics server (existing fire-and-forget behavior is kept).

## Decisions

### D1: Peer label — index instead of truncated public key

**Decision:** Use `strconv.Itoa(i)` (loop index) as the `peer` label value.

**Rationale:** The truncated key is opaque to operators anyway. Index is stable across scrapes (peer order comes from `wgtypes.Device.Peers`, which mirrors config order), shorter, and avoids any partial key exposure concern. Breaking change is acceptable since metrics are new/opt-in and not yet widely deployed.

**Alternative considered:** Config-level `Name` field per peer. Rejected — adds config complexity for a label that peer index handles adequately.

### D2: Sysfs reader as a simple function, not a separate Collector

**Decision:** Add a `ReadIfaceStats(iface string) (IfaceStats, error)` function in `internal/metrics/sysfs.go` that reads `/sys/class/net/<iface>/statistics/{rx_bytes,tx_bytes,...}` via `os.ReadFile`. Called from `Collector.Collect()`.

**Rationale:** Keeping one Collector simplifies registration and ensures all bifrost metrics share one scrape. A separate Collector would need its own `Describe`/`Collect` and complicate the registry setup for no benefit.

**Alternative considered:** Using `net.Interface` or netlink for stats. Rejected — sysfs is simpler, no new dependencies, and the files are always present on Linux.

### D3: Recovery timing via Stats atomics

**Decision:** Add three new atomic fields to `Stats`:
- `recoveryStart` (`atomic.Int64`, unix nano) — set when recovery begins, cleared when it ends.
- `lastRecoveryDuration` (`atomic.Uint64`, float64 bits) — set when recovery completes.
- `lastResolveDuration` (`atomic.Uint64`, float64 bits) — set after each `Resolve()` call.

The supervisor instruments `Recover()` by recording `time.Now()` before and after. `recovery_in_progress` is derived from `recoveryStart != 0`.

**Rationale:** Keeps `Stats` lock-free. The supervisor already calls `Recover()` synchronously, so wrapping it with timing is trivial.

**Alternative considered:** Instrumenting inside `recovery.Recover()` itself. Rejected — `Recover()` is a pure function with injected callbacks; adding a Stats dependency would break its clean interface.

### D4: Tunnel uptime via Stats timestamp

**Decision:** Add `tunnelUpSince` (`atomic.Int64`, unix seconds) to `Stats`. Set to `time.Now().Unix()` in `main.go` after `wg.Bring()` succeeds, and reset by the supervisor after each successful reconnect.

**Rationale:** Unix seconds is the standard for Prometheus `_since` / `_created` gauges. Grafana computes uptime as `time() - bifrost_tunnel_up_since_seconds`.

### D5: Device type as a static CollectorOpts field

**Decision:** Add `DeviceType string` to `CollectorOpts` (value `"kernel"` or `"userspace"`). Expose `Tunnel.IsUserspace() bool` in the wg package.

**Rationale:** Device type is fixed at `Bring()` time and never changes. A static opt is simpler than reading it from `wgtypes.Device` (which doesn't expose it anyway).

### D6: Interface name passed to Collector

**Decision:** Add `Interface string` to `CollectorOpts` so the Collector can read sysfs stats for the correct interface.

## Risks / Trade-offs

- **[Peer label breaking change]** → Acceptable: metrics feature is new (`BIFROST_METRICS` defaults to off) and not yet widely deployed. Document in README.
- **[Sysfs read failure]** → If `/sys/class/net/<iface>/statistics/` is unreadable (container without sysfs mount), interface stats are silently skipped. Other metrics still emit. Log once on first failure.
- **[Recovery timing granularity]** → `Recover()` includes sleep/backoff time in its duration. This is intentional — operators want wall-clock recovery time, not just action time.
- **[Endpoint label cardinality]** → One `bifrost_peer_endpoint_info` series per peer. DDNS changes create new series (old ones go stale). Acceptable for bifrost's typical 1-3 peers.
