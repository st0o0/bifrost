# bifröst

Minimal WireGuard client container with self-healing DDNS recovery, no network namespace teardown on reconnect.

## Architecture

```
cmd/bifrost/          Entry point, healthcheck, handshake CLI, key generation (genkey/pubkey)
internal/
  config/             WireGuard .conf parser + BIFROST_ env var settings
  wg/                 Netlink WireGuard control (kernel-first, wireguard-go fallback)
  recovery/           Two-stage recovery engine (resolve then reconnect), backoff
  supervisor/         Recovery loop: handshake-age staleness + probe trigger
  probe/              ICMP liveness probe (all-down detection), target derivation from AllowedIPs
  metrics/            Prometheus /metrics endpoint (bifrost_ prefix), sysfs interface stats
```

## Build

```bash
go build -o bifrost ./cmd/bifrost
go test ./...
go vet ./...
golangci-lint run
docker build -t bifrost:ci . && ./tests/e2e/run.sh
```

Linux-only packages (`internal/wg`, `cmd/bifrost`) are build-tagged `linux`. Logic packages (`config`, `recovery`, `supervisor`, `probe`, `metrics`) test cross-platform.

## Conventions

- Go 1.26+, kernel-first WireGuard via wgctrl, embedded wireguard-go userspace fallback
- Conventional commits: feat/fix/perf/docs/chore/refactor/test/ci/build/deps
- commitlint enforced, release-please for versioning
- golangci-lint v2, hadolint for Dockerfile
- All env vars prefixed with `BIFROST_`
- Structured logging via log/slog
- Metric names prefixed with `bifrost_`
- Custom Prometheus registry (never DefaultRegisterer)
- Errors wrapped with `fmt.Errorf` and `%w`
- Config loaded via `config.LoadSettings(os.Getenv)` -- testable by injecting getenv

## Deployment

Single static binary in a ~7.5 MB scratch image for linux/amd64 and linux/arm64. Published as `ghcr.io/st0o0/bifrost`. Deployed in homelab-infra as the VPN gateway for dependent containers via `network_mode: service:bifrost`.

## Security-sensitive areas

- `internal/config/` -- parses WireGuard .conf including private keys
- `internal/wg/` -- netlink interface control, key material in memory via wgctrl
- `internal/recovery/` -- interface down/up, endpoint mutation
- `internal/probe/` -- raw ICMP socket (NET_RAW capability)
- `cmd/bifrost/` -- key generation subcommands (genkey, pubkey)
