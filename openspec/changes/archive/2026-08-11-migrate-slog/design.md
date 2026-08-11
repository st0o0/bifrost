## Context

Bifrost uses Go's stdlib `log` package — the only Go service in the fleet still doing so. Eir and ran both use `log/slog` with a near-identical `newLogger()` factory: env-var-driven level + format, stock JSON/Text handlers, no custom `ReplaceAttr`.

Bifrost's logging is concentrated in 4 files (27 log calls total). Core packages (`recovery`, `probe`, `config`) are deliberately log-free, using callbacks instead. This design preserves that.

## Goals / Non-Goals

**Goals:**
- Replace all `log` usage with `slog` across bifrost
- Add `BIFROST_LOG_LEVEL` and `BIFROST_LOG_FORMAT` env vars to `Settings`
- Match the eir/ran `newLogger()` pattern exactly
- Convert printf-style messages to structured key-value attributes
- Default JSON output for Grafana/Loki compatibility

**Non-Goals:**
- Shared logging library across services
- Adding base attributes like `service` or `version` to the logger (follow-up)
- Changing the WireGuard upstream device logger (`device.NewLogger`)
- Changing `health.go` CLI stderr output (`fmt.Fprintln`)
- Async/buffered logging (slog is synchronous, and that's fine)

## Decisions

### 1. Global slog via `slog.SetDefault()` — not dependency injection

Eir passes a `*slog.Logger` explicitly to each component. Ran calls `slog.SetDefault()` so all packages can use the global `slog.Info()` etc.

**Decision: Follow ran's pattern — `slog.SetDefault()`.**

Bifrost's internal packages currently call `log.Printf()` directly (global). Switching to global `slog.Info()` is the minimal diff. Injecting a logger into `wg.Bring()`, `metrics.ListenAndServe()`, and the supervisor would mean changing function signatures and all callers — unnecessary churn for this migration.

### 2. Config: add LogLevel + LogFormat to Settings

Bifrost already has a clean `Settings` struct with `LoadSettings(getenv)`. Add two fields:

```go
LogLevel  slog.Level  // from BIFROST_LOG_LEVEL, default info
LogFormat string      // from BIFROST_LOG_FORMAT, default "json"
```

Use the same `parseLogLevel()` approach as eir/ran (switch on lowercase string → `slog.Level`). Validate `LogFormat` to `"json"` or `"text"` only.

### 3. `log.Fatal` → `slog.Error` + `os.Exit(1)`

`slog` has no `Fatal`. All current `log.Fatal` calls are in `main()` during startup — they abort on unrecoverable config/setup errors. Replace with:

```go
slog.Error("message", "error", err)
os.Exit(1)
```

This is what eir and ran do for the same pattern.

### 4. Log level assignment for existing messages

| Current call | New level | Rationale |
|---|---|---|
| `log.Fatal` (config/setup errors) | `slog.Error` + exit | Unrecoverable |
| `log.Printf` supervisor recovery trigger | `slog.Warn` | Degraded state, action taken |
| `log.Printf` recovery attempt with error | `slog.Warn` | Retry with error |
| `log.Printf` recovery attempt (success) | `slog.Info` | Normal progress |
| `log.Print` "recovered" | `slog.Info` | Positive outcome |
| `log.Print` "recovery exhausted" | `slog.Warn` | Will retry, not terminal |
| `log.Print` "recovery disabled" | `slog.Info` | Informational |
| `log.Printf` WireGuard mode selection | `slog.Info` | Startup info |
| `log.Printf` stale link cleanup | `slog.Warn` | Unexpected state handled |
| `log.Printf` endpoint DNS not resolved | `slog.Warn` | Degraded but continuing |
| `log.Printf` address/route already exists | `slog.Warn` | Idempotency handled |
| `log.Printf` full-tunnel route skip | `slog.Info` | Expected skip |
| `log.Print` probe inactive/on | `slog.Info` | Startup info |
| `log.Printf` metrics server listening | `slog.Info` | Startup info |
| `log.Printf` metrics server error | `slog.Error` | Runtime failure |

### 5. Structured attributes — what to extract from printf strings

Current printf patterns and their slog conversions:

```
log.Printf("%s: using kernel WireGuard", t.iface)
→ slog.Info("using kernel wireguard", "interface", t.iface)

log.Printf("%s attempt %d: %v", stage, a, err)
→ slog.Warn("recovery attempt failed", "stage", stage, "attempt", a, "error", err)

log.Printf("metrics server listening on %s", addr)
→ slog.Info("metrics server listening", "addr", addr)

log.Printf("%s: peer %s endpoint %s:%d does not resolve yet: %v", ...)
→ slog.Warn("peer endpoint not resolved", "interface", iface, "peer", pub, "host", host, "port", port, "error", err)
```

### 6. `newLogger()` placement

In `cmd/bifrost/main.go`, matching eir/ran:

```go
func newLogger(s *Settings) *slog.Logger {
    opts := &slog.HandlerOptions{Level: s.LogLevel}
    if s.LogFormat == "text" {
        return slog.New(slog.NewTextHandler(os.Stdout, opts))
    }
    return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
```

Called after `LoadSettings()`, before any other initialization. This means the first few `log.Fatal` calls (for `LoadSettings` itself) happen before the logger is configured — that's fine, they're fatal startup errors and `slog`'s default text handler covers them.

### 7. Logger initialization timing

```
main()
├── subcommand dispatch (no logging change)
├── LoadSettings()         ← can fail; use slog default (text/info)
├── newLogger(settings)    ← configure structured logger
├── slog.SetDefault()      ← from here on, all slog calls use it
├── LoadConfig()           ← slog.Error + os.Exit(1) on failure
├── wg.Bring()             ← slog calls now structured
├── metrics.ListenAndServe()
└── supervisor.Run()
```

## Risks / Trade-offs

- **Breaking log format**: Users scripting against plain-text `bifrost: ` prefix will break. JSON is the new default. Mitigation: `BIFROST_LOG_FORMAT=text` restores the old-style output.
- **No Fatal flush guarantee**: `slog` is synchronous to `os.Stdout`, so `os.Exit(1)` after `slog.Error()` won't lose the message — stdout is unbuffered by default in Go. No risk here.
- **Supervisor OnAttempt callback still uses global slog**: The callback closure in `supervisor.Run()` will call `slog.Warn`/`slog.Info` directly. Since `slog.SetDefault()` is called before `supervisor.Run()`, this works.
