## MODIFIED Requirements

### Requirement: Peer label format
All per-peer metrics SHALL use a `peer` label containing the zero-based index of the peer as reported by WireGuard (e.g., `"0"`, `"1"`). Peer order is determined by the order peers appear in the `wgtypes.Device.Peers` slice, which mirrors configuration order. The system SHALL NOT expose public keys, endpoint IPs, or hostnames as the primary peer label value.

#### Scenario: Peer label value
- **WHEN** a device has 3 peers
- **THEN** the `peer` label values are `"0"`, `"1"`, `"2"`

#### Scenario: Single peer
- **WHEN** a device has 1 peer
- **THEN** the `peer` label value is `"0"`
