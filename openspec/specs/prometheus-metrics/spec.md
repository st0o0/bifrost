## ADDED Requirements

### Requirement: Metrics endpoint opt-in
The system SHALL expose a Prometheus-compatible `/metrics` HTTP endpoint only when
`BIFROST_METRICS` is set to `on`. The default value SHALL be `off`.

#### Scenario: Metrics disabled by default
- **WHEN** `BIFROST_METRICS` is not set
- **THEN** no HTTP server starts and no port is opened

#### Scenario: Metrics enabled
- **WHEN** `BIFROST_METRICS=on`
- **THEN** an HTTP server starts on the address specified by `BIFROST_METRICS_ADDR`
  serving `/metrics` in Prometheus exposition format

#### Scenario: Invalid listen address
- **WHEN** `BIFROST_METRICS=on` and `BIFROST_METRICS_ADDR` contains an address that
  cannot be bound
- **THEN** the process SHALL exit with an error before entering the supervisor loop

### Requirement: Metrics listen address
The system SHALL accept a `BIFROST_METRICS_ADDR` env var specifying the listen address
for the metrics HTTP server. The default value SHALL be `:9586`.

#### Scenario: Default address
- **WHEN** `BIFROST_METRICS=on` and `BIFROST_METRICS_ADDR` is not set
- **THEN** the HTTP server listens on `:9586`

#### Scenario: Custom address
- **WHEN** `BIFROST_METRICS_ADDR=127.0.0.1:2112`
- **THEN** the HTTP server listens on `127.0.0.1:2112`

### Requirement: Tunnel up gauge
The system SHALL expose a `bifrost_tunnel_up` gauge metric with value `1` when at
least one peer has a handshake newer than the configured `StaleAfter` threshold,
and `0` otherwise.

#### Scenario: Tunnel healthy
- **WHEN** at least one peer has a handshake age less than `StaleAfter`
- **THEN** `bifrost_tunnel_up` equals `1`

#### Scenario: Tunnel stale
- **WHEN** no peer has a handshake newer than `StaleAfter`
- **THEN** `bifrost_tunnel_up` equals `0`

#### Scenario: No handshake yet
- **WHEN** no peer has ever completed a handshake (zero time)
- **THEN** `bifrost_tunnel_up` equals `0`

### Requirement: Peer handshake age gauge
The system SHALL expose a `bifrost_peer_last_handshake_age_seconds` gauge per peer,
reporting the number of seconds since the peer's last handshake.

#### Scenario: Active peer
- **WHEN** a peer last handshaked 45 seconds ago
- **THEN** `bifrost_peer_last_handshake_age_seconds{peer="<first8>"}` equals `45`
  (approximately, within scrape timing)

#### Scenario: Never handshaked
- **WHEN** a peer has never completed a handshake
- **THEN** the metric is not emitted for that peer (or reports 0 with a zero
  handshake time)

### Requirement: Peer transfer counters
The system SHALL expose `bifrost_peer_receive_bytes_total` and
`bifrost_peer_transmit_bytes_total` counter metrics per peer, reporting the
cumulative bytes received and transmitted as reported by WireGuard.

#### Scenario: Traffic flowing
- **WHEN** a peer has received 1048576 bytes and transmitted 524288 bytes
- **THEN** `bifrost_peer_receive_bytes_total{peer="<first8>"}` equals `1048576`
  and `bifrost_peer_transmit_bytes_total{peer="<first8>"}` equals `524288`

### Requirement: Peer label format
All per-peer metrics SHALL use a `peer` label containing the zero-based index of the peer as reported by WireGuard (e.g., `"0"`, `"1"`). Peer order is determined by the order peers appear in the `wgtypes.Device.Peers` slice, which mirrors configuration order. The system SHALL NOT expose public keys, endpoint IPs, or hostnames as the primary peer label value.

#### Scenario: Peer label value
- **WHEN** a device has 3 peers
- **THEN** the `peer` label values are `"0"`, `"1"`, `"2"`

#### Scenario: Single peer
- **WHEN** a device has 1 peer
- **THEN** the `peer` label value is `"0"`

### Requirement: Reconnect counter
The system SHALL expose a `bifrost_reconnects_total` counter metric that increments
by 1 each time the supervisor triggers a reconnect attempt, regardless of success.
The counter SHALL persist for the lifetime of the process.

#### Scenario: After reconnect
- **WHEN** the supervisor has triggered 3 reconnect attempts since startup
- **THEN** `bifrost_reconnects_total` equals `3`

#### Scenario: Counter survives recovery cycles
- **WHEN** a reconnect rebuilds the WireGuard interface
- **THEN** `bifrost_reconnects_total` retains its previous value plus the new attempt

### Requirement: Resolve counter
The system SHALL expose a `bifrost_resolves_total` counter metric that increments
by 1 each time the supervisor triggers a DDNS re-resolution attempt.

#### Scenario: After resolves
- **WHEN** the supervisor has triggered 5 resolve attempts since startup
- **THEN** `bifrost_resolves_total` equals `5`

### Requirement: Endpoint change counter
The system SHALL expose a `bifrost_endpoint_changes_total` counter metric that
increments by 1 each time a DDNS re-resolution yields a different IP address
than the currently configured endpoint.

#### Scenario: Endpoint changed
- **WHEN** a resolve attempt returns IP `203.0.113.5` but the current endpoint is
  `203.0.113.1`
- **THEN** `bifrost_endpoint_changes_total` increments by 1

#### Scenario: Endpoint unchanged
- **WHEN** a resolve attempt returns the same IP as the current endpoint
- **THEN** `bifrost_endpoint_changes_total` does not increment

### Requirement: Probe RTT gauge
The system SHALL expose a `bifrost_probe_rtt_seconds` gauge metric only when
`BIFROST_PROBE=on`. The value SHALL reflect the best (lowest) RTT observed
across all probe targets during the most recent probe tick.

#### Scenario: Probe enabled with reachable target
- **WHEN** `BIFROST_PROBE=on` and the last probe tick observed a best RTT of 12ms
- **THEN** `bifrost_probe_rtt_seconds` equals `0.012`

#### Scenario: Probe disabled
- **WHEN** `BIFROST_PROBE=off` or not set
- **THEN** `bifrost_probe_rtt_seconds` is not registered and not emitted

### Requirement: Build info gauge
The system SHALL expose a `bifrost_build_info` gauge with a `version` label
containing the build version string. The value SHALL always be `1`.

#### Scenario: Version reported
- **WHEN** Bifrost is built with version `0.2.0`
- **THEN** `bifrost_build_info{version="0.2.0"}` equals `1`

### Requirement: Metrics server isolation
The metrics HTTP server SHALL NOT affect tunnel operation. If the HTTP server
encounters an error after startup, it SHALL log the error but SHALL NOT terminate
or stall the supervisor loop.

#### Scenario: Metrics server crash
- **WHEN** the metrics HTTP server goroutine panics or returns an error after
  initial startup
- **THEN** the supervisor loop continues running and the tunnel remains operational

#### Scenario: Scrape under load
- **WHEN** multiple scrapers hit `/metrics` concurrently
- **THEN** the supervisor loop is not blocked or delayed
