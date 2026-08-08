## Why

Bifrost consumers attach via `network_mode: service:bifrost`, making any listener on
the Bifrost container reachable from sibling containers at `127.0.0.1`. One of those
siblings is typically a Grafana Alloy agent that needs to scrape tunnel health metrics.
Today there is no metrics endpoint — operators have no visibility into watchdog activity
(reconnects, DDNS resolves, probe state) without parsing logs.

## What Changes

- Add an opt-in Prometheus `/metrics` HTTP endpoint, gated by `BIFROST_METRICS=on`
  (default `off`, consistent with the existing `BIFROST_PROBE` pattern).
- New `BIFROST_METRICS_ADDR` env var (default `:9586`) controls the listen address.
- Expose `bifrost_`-prefixed metrics sourced from the existing watchdog loop state and
  live `wgctrl` reads — no new polling goroutines.
- New `internal/metrics` package with a `prometheus.Collector` that reads peer state
  on each scrape and a `Stats` struct holding process-lifetime atomic counters.
- New dependency: `github.com/prometheus/client_golang`.

## Non-goals

- Replacing or duplicating `prometheus_wireguard_exporter` naming — Bifrost has its
  own watchdog semantics (reconnects, DDNS, probe) that the community exporter lacks.
- Push-based metrics (Pushgateway, OTLP) — scrape via `/metrics` is sufficient for
  the Alloy sidecar pattern.
- Per-peer endpoint IP/hostname labels — excluded for cardinality and privacy reasons.

## Capabilities

### New Capabilities

- `prometheus-metrics`: Opt-in HTTP server exposing Prometheus metrics for tunnel
  health, peer state, watchdog counters, and build info.

### Modified Capabilities

(none — no existing specs are affected)

## Impact

- **New dependency**: `github.com/prometheus/client_golang` (and transitive deps).
- **`internal/config`**: Two new fields in `Settings` (`Metrics`, `MetricsAddr`) and
  corresponding env-var parsing.
- **`internal/supervisor`**: Callbacks to increment lifetime counters on resolve,
  reconnect, and endpoint-change events. Probe RTT surfaced from `AllDown` results.
- **`internal/recovery`**: Extended `Options` or new callback fields for
  resolve/reconnect/endpoint-change events.
- **`internal/probe`**: `AllDown` or `Pinger` interface extended to return RTT.
- **`cmd/bifrost/main.go`**: Start HTTP server when metrics enabled; fail on bad addr.
- **Docs**: README env-var table, new "Monitoring" section, `.env.example` update.
- **Docker image**: Remains scratch-based; no new ports exposed by default (opt-in).
