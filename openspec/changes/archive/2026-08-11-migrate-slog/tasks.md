## 1. Configuration

- [x] 1.1 Add `LogLevel` (`slog.Level`) and `LogFormat` (`string`) fields to `Settings` in `internal/config/settings.go`, with parsing from `BIFROST_LOG_LEVEL` (default: info) and `BIFROST_LOG_FORMAT` (default: json), including validation
- [x] 1.2 Add tests for the new log settings: defaults, all valid levels, all valid formats, invalid level, invalid format (`internal/config/settings_test.go`)

## 2. Logger setup

- [x] 2.1 Add `newLogger(s *Settings)` function in `cmd/bifrost/main.go` that creates a `*slog.Logger` with JSON or Text handler based on `s.LogFormat` and `s.LogLevel`, call `slog.SetDefault()`, remove `log.SetPrefix` / `log.SetFlags`

## 3. Migrate main.go

- [x] 3.1 Replace all `log.Fatal` / `log.Fatalf` calls in `main()` with `slog.Error()` + `os.Exit(1)`, using structured attributes (`"error"`, `"interface"`, `"path"`)
- [x] 3.2 Replace `log.Print("recovery disabled")` with `slog.Info`, update import from `"log"` to `"log/slog"`

## 4. Migrate supervisor

- [x] 4.1 Replace all `log.Printf` / `log.Print` calls in `internal/supervisor/supervisor.go` with appropriate slog calls (`slog.Warn` for recovery triggers and failed attempts, `slog.Info` for success/status), using structured attributes (`"stage"`, `"attempt"`, `"error"`, `"reason"`, `"targets"`), update import from `"log"` to `"log/slog"`

## 5. Migrate WireGuard device

- [x] 5.1 Replace all `log.Printf` calls in `internal/wg/device.go` with appropriate slog calls (`slog.Info` for mode selection, `slog.Warn` for already-exists/skip/DNS-failure), using structured attributes (`"interface"`, `"peer"`, `"host"`, `"port"`, `"error"`, `"address"`, `"route"`), update import from `"log"` to `"log/slog"`

## 6. Migrate metrics server

- [x] 6.1 Replace `log.Printf` calls in `internal/metrics/server.go` with `slog.Info` (listening) and `slog.Error` (runtime error), using structured attributes (`"addr"`, `"error"`), update import from `"log"` to `"log/slog"`

## 7. Verify

- [x] 7.1 Run `go build ./...` and `go test ./...` to confirm no remaining `"log"` imports and all tests pass
