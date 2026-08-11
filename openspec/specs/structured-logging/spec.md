## Purpose

Structured logging capability for Bifrost, providing configurable log levels, output formats, and structured key-value attributes via Go's `slog` package. Ensures logs are machine-parseable for container runtimes and log collectors.

## Requirements

### Requirement: Configurable log level
The system SHALL support a `BIFROST_LOG_LEVEL` environment variable that controls the minimum severity of emitted log messages. Valid values SHALL be `debug`, `info`, `warn`, and `error` (case-insensitive). The default level SHALL be `info`.

#### Scenario: Default log level
- **WHEN** `BIFROST_LOG_LEVEL` is not set
- **THEN** the system emits messages at `info` level and above, and suppresses `debug` messages

#### Scenario: Debug level enabled
- **WHEN** `BIFROST_LOG_LEVEL` is set to `debug`
- **THEN** the system emits all messages including `debug` level

#### Scenario: Error level only
- **WHEN** `BIFROST_LOG_LEVEL` is set to `error`
- **THEN** the system emits only `error` level messages and suppresses `info` and `warn`

#### Scenario: Invalid log level
- **WHEN** `BIFROST_LOG_LEVEL` is set to an unrecognized value (e.g. `verbose`)
- **THEN** the system SHALL reject the configuration with an error during settings loading

### Requirement: Configurable log format
The system SHALL support a `BIFROST_LOG_FORMAT` environment variable that selects between structured JSON output and human-readable text output. Valid values SHALL be `json` and `text` (case-insensitive). The default format SHALL be `json`.

#### Scenario: Default JSON output
- **WHEN** `BIFROST_LOG_FORMAT` is not set
- **THEN** log messages are emitted as single-line JSON objects containing at minimum `time`, `level`, and `msg` fields

#### Scenario: Text output
- **WHEN** `BIFROST_LOG_FORMAT` is set to `text`
- **THEN** log messages are emitted in slog's key=value text format

#### Scenario: Invalid log format
- **WHEN** `BIFROST_LOG_FORMAT` is set to an unrecognized value (e.g. `xml`)
- **THEN** the system SHALL reject the configuration with an error during settings loading

### Requirement: Structured log attributes
Log messages SHALL include structured key-value attributes instead of interpolated format strings. Domain-specific context (interface name, stage, attempt number, error, address, peer) SHALL be separate attributes, not embedded in the message string.

#### Scenario: Recovery attempt with error
- **WHEN** a recovery stage attempt fails
- **THEN** the log message includes separate `stage`, `attempt`, and `error` attributes

#### Scenario: WireGuard interface setup
- **WHEN** the WireGuard interface is created
- **THEN** the log message includes an `interface` attribute with the interface name

#### Scenario: Metrics server startup
- **WHEN** the metrics server begins listening
- **THEN** the log message includes an `addr` attribute with the listen address

### Requirement: Log output target
The system SHALL write all log output to stdout. This enables container runtimes and log collectors (Docker, Loki/Promtail) to capture logs via the standard stream.

#### Scenario: Logs written to stdout
- **WHEN** the system emits any log message
- **THEN** the message is written to stdout (not stderr)

### Requirement: Fatal error handling
When the system encounters an unrecoverable startup error, it SHALL log the error at `error` level and exit with code 1. The error message SHALL be emitted before the process terminates.

#### Scenario: Config load failure
- **WHEN** settings loading fails (e.g. invalid env var value)
- **THEN** the system logs the error at `error` level with an `error` attribute and exits with code 1

#### Scenario: Interface bring-up failure
- **WHEN** WireGuard interface creation fails
- **THEN** the system logs the error at `error` level and exits with code 1

### Requirement: Metrics server log output
The metrics server SHALL use structured logging via `slog` instead of stdlib `log`. Startup and runtime error messages SHALL include structured attributes.

#### Scenario: Metrics server listening
- **WHEN** the metrics server starts successfully
- **THEN** the system logs at `info` level with `msg` "metrics server listening" and `addr` attribute

#### Scenario: Metrics server runtime error
- **WHEN** the metrics server encounters a runtime error
- **THEN** the system logs at `error` level with an `error` attribute
