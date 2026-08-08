## ADDED Requirements

### Requirement: Tunnel up-since timestamp
The system SHALL expose a `bifrost_tunnel_up_since_seconds` gauge metric containing the Unix timestamp (seconds since epoch) of when the tunnel was last brought up successfully.

#### Scenario: Tunnel brought up
- **WHEN** the WireGuard interface is successfully created at Unix time 1720000000
- **THEN** `bifrost_tunnel_up_since_seconds` equals `1720000000`

#### Scenario: Tunnel reconnected
- **WHEN** a reconnect rebuilds the WireGuard interface at Unix time 1720003600
- **THEN** `bifrost_tunnel_up_since_seconds` equals `1720003600` (reset to the new time)

#### Scenario: Uptime derivation
- **WHEN** the current time is 1720001800 and `bifrost_tunnel_up_since_seconds` is 1720000000
- **THEN** Grafana can compute continuous uptime as `time() - bifrost_tunnel_up_since_seconds` = 1800 seconds

#### Scenario: No tunnel yet
- **WHEN** the tunnel has not been brought up yet (process starting)
- **THEN** `bifrost_tunnel_up_since_seconds` equals `0`
