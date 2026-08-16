## Why

The existing metrics endpoint covers basic tunnel health and peer transfer counters, but operators diagnosing WireGuard issues need deeper visibility: kernel-level interface statistics (drops, errors), recovery episode timing, tunnel uptime tracking, and device/peer metadata. Tools like `wg show` expose much of this data — our Prometheus endpoint should surface it so Grafana dashboards and alerts can use it without shelling out.

## What Changes

- Add `bifrost_device_info` gauge with `type` label (`kernel` / `userspace`) for device metadata.
- Add `bifrost_peer_endpoint_info` gauge with `{peer, endpoint}` labels exposing the peer's current endpoint address.
- Add `bifrost_peer_allowed_ips_count` gauge per peer.
- **BREAKING**: Change peer label strategy from truncated public key (first 8 chars) to peer index (`0`, `1`, ...). All existing per-peer metrics are affected.
- Add 8 interface-level counters from `/sys/class/net/<iface>/statistics/`: rx/tx bytes, packets, errors, drops.
- Add recovery observability: `bifrost_recovery_duration_seconds`, `bifrost_recovery_in_progress`, `bifrost_resolve_duration_seconds`.
- Add `bifrost_tunnel_up_since_seconds` gauge (unix timestamp) for continuous uptime tracking.

## Non-goals

- Exposing peer public keys or preshared key material in metrics.
- Per-peer endpoint history or cardinality tracking.
- Push-based metrics (Pushgateway, OTLP) — pull via `/metrics` only.
- Graceful shutdown of the metrics HTTP server (existing behavior unchanged).

## Capabilities

### New Capabilities
- `interface-stats`: Kernel-level interface statistics from sysfs (rx/tx bytes, packets, errors, drops).
- `recovery-observability`: Recovery episode timing, in-progress flag, and resolve duration metrics.
- `tunnel-uptime`: Continuous tunnel uptime tracking via unix timestamp gauge.
- `device-metadata`: Device type info gauge and peer endpoint/allowed-IPs metadata.

### Modified Capabilities
- `prometheus-metrics`: Peer label format changes from truncated public key to peer index. New per-peer metadata metrics added.

## Impact

- **`internal/metrics/`**: `collector.go` and `stats.go` gain new fields and descriptors. New `sysfs.go` for interface stats.
- **`internal/supervisor/supervisor.go`**: Instruments recovery calls with timing.
- **`internal/wg/`**: Exposes device type (userspace bool) to collector.
- **`cmd/bifrost/main.go`**: Passes new options (device type, interface name) to collector.
- **Existing dashboards**: Peer label value changes from key prefix to index — queries using `{peer="..."}` must be updated.
- **No new dependencies**: sysfs reading uses `os.ReadFile`, no new Go modules.
