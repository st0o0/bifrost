## 1. Configuration

- [x] 1.1 Add `Metrics` (bool) and `MetricsAddr` (string) fields to `internal/config/settings.go` with env var parsing for `BIFROST_METRICS` and `BIFROST_METRICS_ADDR`
- [x] 1.2 Add unit tests for the new settings fields in `internal/config/settings_test.go` (defaults, overrides, validation)

## 2. Stats struct

- [x] 2.1 Create `internal/metrics/stats.go` with the `Stats` struct: atomic counters for reconnects, resolves, endpoint changes, and probe RTT; `Inc*()` and `Set*()` methods
- [x] 2.2 Add unit tests for `Stats` in `internal/metrics/stats_test.go` (concurrent increments, RTT set/get)

## 3. Probe RTT plumbing

- [x] 3.1 Change `probe.AllDown()` to return a `Result` struct (`Down bool`, `BestRTT time.Duration`) instead of a plain `bool`; update `alldown.go` and callers in `internal/supervisor/supervisor.go`
- [x] 3.2 Update `internal/probe/alldown_test.go` and `internal/supervisor/supervisor_test.go` for the new return type

## 4. Recovery callbacks

- [x] 4.1 Add `OnResolve`, `OnReconnect`, `OnEndpointChange` callback fields to `recovery.Options` in `internal/recovery/recovery.go`; invoke them from `Recover()`
- [x] 4.2 Wire callbacks in `internal/supervisor/supervisor.go` to call `Stats.IncReconnects()`, `Stats.IncResolves()`, `Stats.IncEndpointChanges()`
- [x] 4.3 Add endpoint change detection in `internal/wg/control.go` `Resolve()`: compare old vs new resolved IP, return a bool indicating change
- [x] 4.4 Update `internal/recovery/recovery_test.go` to verify callbacks fire

## 5. Prometheus Collector

- [x] 5.1 Create `internal/metrics/collector.go` implementing `prometheus.Collector` with a `DeviceFunc` dependency for wgctrl reads; emit all specified metrics with `bifrost_` prefix
- [x] 5.2 Add unit tests in `internal/metrics/collector_test.go` with a stubbed `DeviceFunc` returning fake peer state

## 6. HTTP server and main wiring

- [x] 6.1 Create `internal/metrics/server.go` with a function to start the HTTP server (listener + `/metrics` handler); fail-fast on bind error
- [x] 6.2 Wire everything in `cmd/bifrost/main.go`: create `Stats`, create and register `Collector`, start metrics server when `settings.Metrics` is true, pass `Stats` to supervisor

## 7. Build info

- [x] 7.1 Add `bifrost_build_info` gauge with `version` label to the Collector using `debug.ReadBuildInfo()` (in `internal/metrics/collector.go`)

## 8. Documentation

- [x] 8.1 Update `README.md`: add `BIFROST_METRICS` and `BIFROST_METRICS_ADDR` to the env-var table; add a "Monitoring" section with an example Alloy `prometheus.scrape` block
- [x] 8.2 Update `.env.example` with the two new env vars and comments
