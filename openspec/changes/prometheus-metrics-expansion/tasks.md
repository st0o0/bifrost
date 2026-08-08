## 1. Peer label change & device metadata

- [ ] 1.1 Change `peerLabel()` in `collector.go` to use peer index (`strconv.Itoa(i)`) instead of truncated public key
- [ ] 1.2 Add `DeviceType string` and `Interface string` to `CollectorOpts`; add `bifrost_device_info` desc and emit in `Collect()`
- [ ] 1.3 Expose `Tunnel.IsUserspace() bool` in `internal/wg/device.go`
- [ ] 1.4 Update `cmd/bifrost/main.go` to pass `DeviceType` and `Interface` to `CollectorOpts`
- [ ] 1.5 Update `collector_test.go`: adjust peer label expectations to index, add test for `bifrost_device_info`

## 2. Peer endpoint & allowed IPs metadata

- [ ] 2.1 Add `bifrost_peer_endpoint_info` and `bifrost_peer_allowed_ips_count` descs to `Collector`; emit in the per-peer loop in `Collect()`
- [ ] 2.2 Add tests for endpoint info gauge (peer with/without endpoint) and allowed IPs count in `collector_test.go`

## 3. Interface sysfs statistics

- [ ] 3.1 Create `internal/metrics/sysfs.go` (Linux build-tagged): `IfaceStats` struct and `ReadIfaceStats(iface string) (IfaceStats, error)` reading from `/sys/class/net/<iface>/statistics/`
- [ ] 3.2 Create `internal/metrics/sysfs_test.go`: test with a temp directory simulating sysfs layout, test missing/partial files
- [ ] 3.3 Add 8 interface counter descs to `Collector`; call `ReadIfaceStats()` in `Collect()` and emit when available
- [ ] 3.4 Add collector test verifying interface metrics are emitted when sysfs data is available and omitted when not

## 4. Recovery observability

- [ ] 4.1 Add `recoveryStart`, `lastRecoveryDuration`, `lastResolveDuration` atomic fields to `Stats` with getter/setter methods
- [ ] 4.2 Add tests for new `Stats` fields in `stats_test.go`
- [ ] 4.3 Instrument `supervisor.Run()`: set `recoveryStart` before calling `Recover()`, compute and store duration after, time `Resolve()` calls
- [ ] 4.4 Add `bifrost_recovery_duration_seconds`, `bifrost_recovery_in_progress`, `bifrost_resolve_duration_seconds` descs to `Collector`; emit in `Collect()`
- [ ] 4.5 Add collector tests for recovery metrics (in-progress, duration, resolve duration)

## 5. Tunnel uptime

- [ ] 5.1 Add `tunnelUpSince` atomic field to `Stats` with `SetTunnelUpSince(unix int64)` and `TunnelUpSince() int64`
- [ ] 5.2 Set `tunnelUpSince` in `main.go` after `wg.Bring()` succeeds; reset in supervisor after successful reconnect
- [ ] 5.3 Add `bifrost_tunnel_up_since_seconds` desc to `Collector`; emit in `Collect()`
- [ ] 5.4 Add collector and stats tests for tunnel uptime metric

## 6. Documentation

- [ ] 6.1 Update `.env.example` with `BIFROST_METRICS` and `BIFROST_METRICS_ADDR` entries
- [ ] 6.2 Update `README.md` with full metrics table and peer label change note
