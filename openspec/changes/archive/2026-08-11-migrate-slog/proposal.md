## Why

Bifrost is the last Go service in the fleet (alongside eir and ran) still using Go's stdlib `log` package. This means no structured output, no log levels, no JSON, and no way to filter verbosity — making it the odd one out when building unified Grafana/Loki dashboards across all services.

## What Changes

- Replace all `log.Printf` / `log.Fatal` calls with `slog.Info` / `slog.Error` / `slog.Warn` / `slog.Debug`
- Add `newLogger()` factory matching the eir/ran pattern (env-var driven level + format)
- Add `BIFROST_LOG_LEVEL` and `BIFROST_LOG_FORMAT` configuration options
- Convert printf-style format strings to structured key-value attributes
- Default to JSON output for Grafana/Loki consumption

## Non-goals

- Changing the logging approach in core packages (`recovery`, `probe`) — they stay log-free via callbacks
- Modifying the upstream WireGuard device logger (`device.NewLogger`)
- Changing the CLI stderr output in `health.go` (`fmt.Fprintln`)
- Adding a shared logging library across services — each service keeps its own `newLogger()`
- Adding service-wide base attributes (`service`, `version`) — that's a follow-up concern

## Capabilities

### New Capabilities

- `structured-logging`: Configurable structured logging via `log/slog` with level filtering and JSON/text output format selection

### Modified Capabilities

- `prometheus-metrics`: The metrics server log calls move from `log.Printf` to `slog`

## Impact

- **Code**: 4 files across 3 packages (`cmd/bifrost`, `internal/supervisor`, `internal/wg`, `internal/metrics`)
- **Config**: 2 new env vars (`BIFROST_LOG_LEVEL`, `BIFROST_LOG_FORMAT`) added to `internal/config`
- **Dependencies**: None — `log/slog` is stdlib since Go 1.21
- **Breaking**: Log output format changes from plain text to structured JSON by default. Users parsing log output will need to adapt.
