## Purpose

Expose Prometheus metrics for observing the supervisor's recovery and resolve operations, including duration tracking and in-progress status.

## Requirements

### Requirement: Recovery duration gauge
The system SHALL expose a `bifrost_recovery_duration_seconds` gauge metric reporting the wall-clock duration of the most recently completed recovery episode in seconds. The value SHALL include all retry attempts and backoff sleeps within the episode.

#### Scenario: Recovery completed
- **WHEN** a recovery episode started at T and completed successfully at T+45s
- **THEN** `bifrost_recovery_duration_seconds` equals `45`

#### Scenario: Recovery failed
- **WHEN** a recovery episode started at T and exhausted all retries at T+120s without recovering
- **THEN** `bifrost_recovery_duration_seconds` equals `120`

#### Scenario: No recovery yet
- **WHEN** no recovery episode has occurred since process start
- **THEN** `bifrost_recovery_duration_seconds` equals `0`

### Requirement: Recovery in-progress flag
The system SHALL expose a `bifrost_recovery_in_progress` gauge metric with value `1` when a recovery episode is currently executing, and `0` otherwise.

#### Scenario: Recovery running
- **WHEN** the supervisor is currently executing a recovery episode
- **THEN** `bifrost_recovery_in_progress` equals `1`

#### Scenario: No recovery running
- **WHEN** no recovery episode is in progress
- **THEN** `bifrost_recovery_in_progress` equals `0`

#### Scenario: Recovery finishes
- **WHEN** a recovery episode completes (success or exhaustion)
- **THEN** `bifrost_recovery_in_progress` transitions from `1` to `0`

### Requirement: Resolve duration gauge
The system SHALL expose a `bifrost_resolve_duration_seconds` gauge metric reporting the wall-clock duration of the most recent `Resolve()` call in seconds.

#### Scenario: Resolve completed
- **WHEN** the last `Resolve()` call took 0.250 seconds
- **THEN** `bifrost_resolve_duration_seconds` equals `0.25`

#### Scenario: No resolve yet
- **WHEN** no `Resolve()` call has occurred since process start
- **THEN** `bifrost_resolve_duration_seconds` equals `0`
