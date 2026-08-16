## ADDED Requirements

### Requirement: Device info gauge
The system SHALL expose a `bifrost_device_info` gauge with value `1` and a `type` label indicating whether the WireGuard device is using the kernel module (`"kernel"`) or the userspace implementation (`"userspace"`).

#### Scenario: Kernel WireGuard
- **WHEN** the WireGuard interface was created using the kernel module
- **THEN** `bifrost_device_info{type="kernel"}` equals `1`

#### Scenario: Userspace WireGuard
- **WHEN** the WireGuard interface fell back to the userspace (wireguard-go) implementation
- **THEN** `bifrost_device_info{type="userspace"}` equals `1`

### Requirement: Peer endpoint info gauge
The system SHALL expose a `bifrost_peer_endpoint_info` gauge with value `1` per peer, with labels `{peer, endpoint}` where `endpoint` is the peer's current endpoint address as reported by WireGuard (format `ip:port`).

#### Scenario: Peer with endpoint
- **WHEN** peer 0 has endpoint `203.0.113.1:51820`
- **THEN** `bifrost_peer_endpoint_info{peer="0", endpoint="203.0.113.1:51820"}` equals `1`

#### Scenario: Peer without endpoint
- **WHEN** a peer has no endpoint configured or resolved yet
- **THEN** `bifrost_peer_endpoint_info` is not emitted for that peer

#### Scenario: DDNS endpoint change
- **WHEN** peer 0's endpoint changes from `203.0.113.1:51820` to `203.0.113.5:51820`
- **THEN** `bifrost_peer_endpoint_info{peer="0", endpoint="203.0.113.5:51820"}` equals `1`
- **AND** the old series `{endpoint="203.0.113.1:51820"}` goes stale naturally

### Requirement: Peer allowed IPs count gauge
The system SHALL expose a `bifrost_peer_allowed_ips_count` gauge per peer reporting the number of allowed IP prefixes configured for that peer.

#### Scenario: Peer with multiple allowed IPs
- **WHEN** peer 0 has 3 allowed IP prefixes
- **THEN** `bifrost_peer_allowed_ips_count{peer="0"}` equals `3`

#### Scenario: Peer with no allowed IPs
- **WHEN** peer 1 has 0 allowed IP prefixes
- **THEN** `bifrost_peer_allowed_ips_count{peer="1"}` equals `0`
