## Purpose

Expose WireGuard network interface statistics from the kernel's sysfs as Prometheus counter metrics, with graceful degradation when sysfs is unavailable.

## Requirements

### Requirement: Interface receive bytes counter
The system SHALL expose a `bifrost_interface_rx_bytes_total` counter metric reporting the cumulative bytes received on the WireGuard network interface, as read from the kernel's sysfs statistics.

#### Scenario: Bytes received
- **WHEN** the kernel reports 2097152 received bytes on the WireGuard interface
- **THEN** `bifrost_interface_rx_bytes_total` equals `2097152`

#### Scenario: Sysfs unavailable
- **WHEN** `/sys/class/net/<iface>/statistics/rx_bytes` is not readable
- **THEN** the metric is not emitted and no error is returned to the scraper

### Requirement: Interface transmit bytes counter
The system SHALL expose a `bifrost_interface_tx_bytes_total` counter metric reporting the cumulative bytes transmitted on the WireGuard network interface.

#### Scenario: Bytes transmitted
- **WHEN** the kernel reports 1048576 transmitted bytes on the WireGuard interface
- **THEN** `bifrost_interface_tx_bytes_total` equals `1048576`

### Requirement: Interface receive packets counter
The system SHALL expose a `bifrost_interface_rx_packets_total` counter metric reporting the cumulative packets received on the WireGuard network interface.

#### Scenario: Packets received
- **WHEN** the kernel reports 15000 received packets
- **THEN** `bifrost_interface_rx_packets_total` equals `15000`

### Requirement: Interface transmit packets counter
The system SHALL expose a `bifrost_interface_tx_packets_total` counter metric reporting the cumulative packets transmitted on the WireGuard network interface.

#### Scenario: Packets transmitted
- **WHEN** the kernel reports 12000 transmitted packets
- **THEN** `bifrost_interface_tx_packets_total` equals `12000`

### Requirement: Interface receive errors counter
The system SHALL expose a `bifrost_interface_rx_errors_total` counter metric reporting the cumulative receive errors on the WireGuard network interface.

#### Scenario: Receive errors
- **WHEN** the kernel reports 3 receive errors
- **THEN** `bifrost_interface_rx_errors_total` equals `3`

### Requirement: Interface transmit errors counter
The system SHALL expose a `bifrost_interface_tx_errors_total` counter metric reporting the cumulative transmit errors on the WireGuard network interface.

#### Scenario: Transmit errors
- **WHEN** the kernel reports 1 transmit error
- **THEN** `bifrost_interface_tx_errors_total` equals `1`

### Requirement: Interface receive dropped counter
The system SHALL expose a `bifrost_interface_rx_dropped_total` counter metric reporting the cumulative dropped received packets on the WireGuard network interface.

#### Scenario: Packets dropped on receive
- **WHEN** the kernel reports 5 dropped received packets
- **THEN** `bifrost_interface_rx_dropped_total` equals `5`

### Requirement: Interface transmit dropped counter
The system SHALL expose a `bifrost_interface_tx_dropped_total` counter metric reporting the cumulative dropped transmitted packets on the WireGuard network interface.

#### Scenario: Packets dropped on transmit
- **WHEN** the kernel reports 2 dropped transmitted packets
- **THEN** `bifrost_interface_tx_dropped_total` equals `2`

### Requirement: Graceful sysfs degradation
The system SHALL NOT fail or return an error to the Prometheus scraper if sysfs statistics are unavailable (e.g., container without sysfs mount). When unavailable, all interface statistics metrics SHALL be omitted silently.

#### Scenario: Sysfs not mounted
- **WHEN** `/sys/class/net/<iface>/statistics/` does not exist
- **THEN** no `bifrost_interface_*` metrics are emitted
- **AND** all other metrics continue to be emitted normally

#### Scenario: Partial sysfs availability
- **WHEN** some sysfs statistics files exist but others do not
- **THEN** only the readable statistics are emitted as metrics
