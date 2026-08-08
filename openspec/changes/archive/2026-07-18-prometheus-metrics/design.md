## Context

Bifrost is a minimal WireGuard sidecar that runs a supervisor loop reading peer state
via `wgctrl` and triggering resolve/reconnect recovery when handshakes go stale or
liveness probes fail. It has no HTTP server, no metrics, and no external monitoring
surface beyond Docker healthcheck exit codes and log output.

Sibling containers sharing Bifrost's network namespace (via `network_mode: service:bifrost`)
can reach any listener at `127.0.0.1`. A Grafana Alloy agent in this position needs a
Prometheus scrape target to collect tunnel health metrics.

## Goals / Non-Goals

**Goals:**

- Expose Bifrost watchdog state as Prometheus metrics on an opt-in HTTP endpoint.
- Source all metric values from existing loop state and live `wgctrl` reads — no new
  polling goroutines.
- Keep the metrics server completely decoupled from tunnel operation — a metrics
  failure must never crash or stall the supervisor.
- Follow the project's style: minimal dependencies, dependency injection, testable
  without Linux or real WireGuard.

**Non-Goals:**

- Push-based telemetry (OTLP, Pushgateway).
- Replacing the Docker healthcheck with an HTTP liveness endpoint.
- Histogram-type metrics (scrape latency distributions, probe RTT percentiles).

## Decisions

### 1. `prometheus.Collector` interface over pre-registered gauges

**Decision**: Implement `prometheus.Collector` on a custom struct. On each `Collect()`
call, query `wgctrl` for live peer state and read atomic counters from the shared
`Stats` struct.

**Alternatives considered**:
- *Update gauges in the supervisor loop*: Would make metrics stale between ticks
  (30s default). Transfer byte counters would only update on check intervals, not on
  scrape. Also couples Prometheus types into the supervisor.
- *expvar + a Prometheus adapter*: Adds an unnecessary translation layer.

**Rationale**: The Collector approach reads `wgctrl` on demand (a cheap netlink ioctl),
so metrics are always fresh at scrape time. The supervisor stays unaware of Prometheus.

### 2. Shared `Stats` struct with atomic counters

**Decision**: A `metrics.Stats` struct holds `uint64` atomic counters for reconnects,
resolves, and endpoint changes, plus a `float64` atomic for probe RTT. The supervisor
increments these via method calls; the Collector reads them.

**Alternatives considered**:
- *Channel-based event bus*: Over-engineered for simple counters.
- *Prometheus counters directly in the supervisor*: Couples the supervisor to Prometheus
  types; harder to test without the full registry.

**Rationale**: Atomics are lock-free, trivially testable, and keep the Prometheus
dependency contained in `internal/metrics`.

### 3. Probe RTT surfaced via extended `AllDown` return

**Decision**: Change `probe.AllDown()` to return a `Result` struct containing both
the `down bool` and the best RTT observed across targets. The supervisor stores this
in `Stats.SetProbeRTT()`. The metric is only registered when `BIFROST_PROBE=on`.

**Alternatives considered**:
- *Collector runs its own ping on scrape*: Adds latency to scrape, could interfere
  with probe timing, duplicates ICMP logic.
- *Extend the `Pinger` interface*: Would require changing every `Pinger` implementation
  (including test fakes) for a single float.

**Rationale**: `AllDown` already iterates all targets and gets `Statistics()` back from
`pro-bing`. Returning the RTT alongside the bool is a minimal change.

### 4. Endpoint change detection in `Resolve()`

**Decision**: `wg.Resolve()` already re-resolves the DDNS hostname and calls
`wgctrl.ConfigureDevice` with the new endpoint. We add a comparison of old vs new
resolved IP before applying — if they differ, the supervisor increments
`Stats.IncEndpointChanges()`.

**Alternatives considered**:
- *Detect in the Collector by caching last-seen endpoint*: The Collector shouldn't
  track state across scrapes; that's the supervisor's job.

### 5. Peer label: first 8 chars of base64 public key

**Decision**: Use `peer="a1B2c3D4"` (first 8 characters of the base64-encoded public
key). Never expose endpoint IPs or hostnames.

**Rationale**: 8 chars provides 48 bits of entropy — collision-free for any realistic
peer count. Full keys are unnecessarily long for dashboard labels. IPs/hostnames are
excluded for cardinality and privacy.

### 6. `tunnel_up` semantics

**Decision**: `bifrost_tunnel_up` = 1 if any peer has a handshake newer than
`StaleAfter` seconds. This matches the supervisor's own definition of liveness.

### 7. HTTP server lifecycle

**Decision**: Start `net/http.Server` in a goroutine from `main.go`, after
`wg.Bring()` succeeds. If `net.Listen` fails on the configured address, log the
error and exit immediately (fail-fast, before entering the supervisor loop).
The server runs independently; if it panics or errors, log but do not terminate
the supervisor.

## Risks / Trade-offs

- **[New dependency]** `prometheus/client_golang` pulls in `protobuf`, `procfs`, etc.
  → Acceptable: these are widely used, well-maintained, and only linked when metrics
  are enabled. Binary size increase is ~2-3 MB.

- **[wgctrl call on every scrape]** A misbehaving scraper with a very short interval
  could cause frequent netlink calls. → Mitigation: `wgctrl.Device()` is a single
  ioctl, sub-millisecond. Even at 1s scrape intervals this is negligible.

- **[Probe RTT only updates on probe tick]** The RTT gauge shows the last observed
  value from the probe interval (default 10s), not a live measurement. → Acceptable:
  this matches what the watchdog actually sees. Documentation will clarify.

- **[Counter reset on process restart]** Atomic counters reset to zero on restart.
  → Expected Prometheus behavior; `rate()` / `increase()` handle resets correctly.
