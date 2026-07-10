# bifrost

[![CI](https://github.com/st0o0/bifrost/actions/workflows/ci.yml/badge.svg)](https://github.com/st0o0/bifrost/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/st0o0/bifrost?sort=semver)](https://github.com/st0o0/bifrost/releases)
[![GHCR](https://img.shields.io/badge/ghcr.io-st0o0%2Fbifrost-2496ED?logo=docker&logoColor=white)](https://github.com/st0o0/bifrost/pkgs/container/bifrost)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A minimal WireGuard **client** container that keeps dependent containers
tunneled across DDNS endpoint IP changes — **without tearing down its network
namespace**. The Norse rainbow bridge between your networks.

```yaml
services:
  bifrost:
    image: ghcr.io/st0o0/bifrost:latest
    cap_add: [NET_ADMIN]
    volumes:
      - ./wg/wg0.conf:/etc/wireguard/wg0.conf:ro
    ports:
      - "127.0.0.1:8200:8200"   # a port of the app tunneled below

  app:
    image: yourapp:latest
    network_mode: "service:bifrost"   # rides the tunnel
    depends_on:
      bifrost:
        condition: service_healthy
```

## Features

- **Self-healing, two-stage recovery.** A supervisor watches handshake
  freshness and, on a stale tunnel, first re-resolves the endpoint in place
  (`wg set`, no interface flap), then — only if that fails — actively rebuilds
  the tunnel (`wg-quick down/up`). Both stages retry with exponential backoff.
- **DDNS-native.** The endpoint may be a hostname; bifrost re-resolves it when
  the peer's IP changes. Dependents on `network_mode: service:bifrost` never
  lose their network namespace during recovery.
- **Kernel-first, userspace fallback.** Uses the in-kernel WireGuard when
  available; otherwise a bundled `wireguard-go`. No `SYS_MODULE`, no
  `/lib/modules` mount, no PUID/PGID, no s6.
- **Split-tunnel friendly.** No forced killswitch — route only the subnets your
  `AllowedIPs` list, everything else goes direct.
- **Fully configurable & toggleable.** Every threshold, retry count, and backoff
  is an env var; the resolve stage, the reconnect stage, and the healthcheck can
  each be turned off.
- **Small & multi-arch.** ~30 MB Alpine image for `linux/amd64` and
  `linux/arm64`.

## Why

| | linuxserver/wireguard | gluetun | **bifrost** |
|---|:---:|:---:|:---:|
| Runs as a client that others route through | ✅ | ✅ | ✅ |
| DDNS hostname endpoint | ⚠️ no re-resolve | ❌ IP only | ✅ |
| In-place recovery (no netns teardown) | ❌ | ⚠️ tunnel flap | ✅ |
| Active reconnect on a dead tunnel | ❌ | ✅ | ✅ |
| Split-tunnel without a killswitch | ✅ | ❌ full-tunnel | ✅ |
| No kernel-module mount / PUID / s6 | ❌ | ✅ | ✅ |

- **linuxserver/wireguard** needs the kernel module mounted (`SYS_MODULE`,
  `/lib/modules`), s6, PUID/PGID; recovering means a container restart, which
  drops every container routed through it.
- **gluetun** is a full-tunnel killswitch and [cannot use a DDNS hostname
  endpoint](https://github.com/qdm12/gluetun/issues/2680) (IP only).

## Quick start

1. Put your WireGuard client config at `./wg/wg0.conf`. An existing
   linuxserver/wireguard client config works unchanged. For NAT keepalive add
   `PersistentKeepalive = 25` under `[Peer]`.
2. `cp .env.example .env` (optional — every value has a sane default).
3. `docker compose up -d`

```bash
docker logs -f bifrost                                       # data path + recovery
docker inspect --format '{{.State.Health.Status}}' bifrost   # healthy / unhealthy
```

Publish the ports of any tunneled service on the **bifrost** container, and give
the service `network_mode: "service:bifrost"`.

### Split tunnel (reach a few internal hosts)

Route only specific hosts/subnets through the tunnel; everything else stays
direct. This is driven entirely by your `.conf`:

```ini
[Interface]
PrivateKey = <client key>
Address    = 10.13.13.2/32

[Peer]
PublicKey           = <server key>
Endpoint            = vpn.example.com:51820   # DDNS hostname is fine
AllowedIPs          = 10.50.0.10/32, 10.50.0.11/32, 10.13.13.1/32
PersistentKeepalive = 25
```

### Full tunnel

Set `AllowedIPs = 0.0.0.0/0, ::/0`. bifrost does **not** add a killswitch; if you
want to block leaks when the tunnel is down, add firewall rules yourself.

## How recovery works

bifrost never relies on a container restart. The supervisor loop checks the
newest handshake every `BIFROST_CHECK_INTERVAL`. Once it is older than
`BIFROST_STALE_AFTER`, a **recovery episode** starts and escalates:

```
handshake age >  STALE_AFTER  ──▶  Stage 1: RESOLVE   (wg set endpoint, ×RETRIES, backoff)
        │  recovered? ──────────▶  done
        ▼  exhausted
                                   Stage 2: RECONNECT (wg-quick down/up, ×RETRIES, backoff)
        │  recovered? ──────────▶  done
        ▼  exhausted
                                   log + keep monitoring (re-tries next tick)
```

Recovery counts as successful only when a **new** handshake occurs (strictly
newer than the one at episode start), so a stale endpoint can never look
healthy. Backoff is exponential (`BACKOFF · 2ⁿ`, capped at 60 s). Either stage
can be disabled; with both off, the supervisor doesn't run.

Default escalation ordering: **resolve at 135 s** → **healthcheck flips at
180 s** → **reconnect** after the resolve retries elapse.

## Configuration

All configuration is via environment variables (see `.env.example`).

### Liveness probe (optional, faster detection)

Recovery normally triggers on handshake age, whose floor is ~135 s (WireGuard
rekeys a healthy tunnel roughly every 120 s, and its timers aren't tunable). To
react faster, enable an active ICMP probe: it pings the tunnel-internal hosts
and triggers recovery when they go unreachable — typically within ~30 s at the
conservative defaults. It is an **additional, independent** trigger; whichever
fires first (probe or handshake age) starts recovery.

| Variable | Default | Purpose |
|---|---|---|
| `BIFROST_PROBE` | `off` | Enable the liveness probe |
| `BIFROST_PROBE_INTERVAL` | `10` | Seconds per probe round |
| `BIFROST_PROBE_FAILS` | `3` | Consecutive "all targets down" rounds before triggering |
| `BIFROST_PROBE_TIMEOUT` | `2` | Per-ping timeout (seconds) |
| `BIFROST_PROBE_HOST` | *(AllowedIPs)* | Explicit target(s); default derives `/32`+`/128` hosts from AllowedIPs |

Targets are pinged **inside** the tunnel; a round counts as down only when
**every** target fails, so a single offline host never triggers recovery — only
a genuinely dead tunnel does. Ranges and `0.0.0.0/0` are skipped (not pingable);
for a full-tunnel setup set `BIFROST_PROBE_HOST` explicitly. At least one target
must answer ICMP — otherwise every round reads "all down"
and (with the reconnect stage active) flaps a healthy tunnel. On first enable,
set `BIFROST_PROBE_HOST` to your server's tunnel IP, which reliably answers.

### Interface & monitoring

| Variable | Default | Purpose |
|---|---|---|
| `BIFROST_INTERFACE` | `wg0` | Config/interface name → `/etc/wireguard/<name>.conf` |
| `BIFROST_CHECK_INTERVAL` | `30` | How often to check handshake freshness (seconds, ≥ 1) |
| `BIFROST_STALE_AFTER` | `135` | Handshake age (seconds) that triggers a recovery episode |

### Recovery — stage 1: resolve (gentle, in place)

| Variable | Default | Purpose |
|---|---|---|
| `BIFROST_RESOLVE` | `on` | Enable the resolve stage |
| `BIFROST_RESOLVE_RETRIES` | `5` | Attempts per episode |
| `BIFROST_RESOLVE_BACKOFF` | `5` | Base backoff (seconds, exponential, cap 60) |

### Recovery — stage 2: reconnect (active `wg-quick down/up`)

| Variable | Default | Purpose |
|---|---|---|
| `BIFROST_RECONNECT` | `on` | Enable the reconnect stage |
| `BIFROST_RECONNECT_RETRIES` | `5` | Attempts per episode |
| `BIFROST_RECONNECT_BACKOFF` | `5` | Base backoff (seconds, exponential, cap 60) |

### Healthcheck

| Variable | Default | Purpose |
|---|---|---|
| `BIFROST_HEALTHCHECK` | `on` | Enable the container healthcheck |
| `BIFROST_HEALTH_STALE_AFTER` | `180` | Report unhealthy if the handshake is older (seconds) |

The image ships a Docker `HEALTHCHECK`. The **schedule** (`--interval`,
`--timeout`, `--retries`) is baked at build time; override it per deployment
with a compose `healthcheck:` block. Toggles accept `on/off/true/false/yes/no/1/0`;
invalid or non-numeric values fail fast at startup with a clear message.

> Setting `BIFROST_HEALTHCHECK=off` makes the container report healthy
> unconditionally, so a dependent using `depends_on: condition: service_healthy`
> starts regardless of tunnel state.

If you change `BIFROST_INTERFACE` away from `wg0`, mount your config at the
matching `/etc/wireguard/<name>.conf`.

## Runtime requirements

WireGuard runs either **in the kernel** (fast, no TUN device) or **in userspace**
(needs `/dev/net/tun`). What a container needs depends on which path it uses and
whether it loads the kernel module itself:

| Image | `NET_ADMIN` | `SYS_MODULE` | `/lib/modules` mount | `/dev/net/tun` | Data path |
|---|:---:|:---:|:---:|:---:|---|
| **bifrost** | ✅ | — | — | only w/o kernel module | kernel-first, userspace fallback |
| linuxserver/wireguard | ✅ | ✅ | ✅ | — | kernel module (loads it) |
| gluetun | ✅ | — | — | ✅ | userspace (TUN) |
| wg-easy *(server)* | ✅ | ✅ | ✅ | — | kernel module |

**When do I need `/dev/net/tun`?** Only when WireGuard runs in userspace. For
bifrost that's *only* if the host has no in-kernel WireGuard module. WireGuard
has shipped in the Linux kernel since 5.6 (2020), so on virtually any modern host
**bifrost needs just `cap_add: [NET_ADMIN]`** — nothing else. Add `/dev/net/tun`
for old kernels, some NAS boxes, or restricted/managed hosts (harmless to always
include).

**When do I need `SYS_MODULE` + `/lib/modules`?** Only for images that (re)load
the kernel module themselves (linuxserver, wg-easy). bifrost never does — it uses
the module if present, otherwise it falls back to userspace.

Quick check — does your host already have in-kernel WireGuard?

```bash
if test -d /sys/module/wireguard || modprobe wireguard 2>/dev/null; then
  echo "in-kernel WireGuard available — bifrost needs only NET_ADMIN"
else
  echo "no kernel module — add --device /dev/net/tun for the userspace fallback"
fi
```

## Image

- Registry: `ghcr.io/st0o0/bifrost`
- Tags: `latest`, `MAJOR.MINOR`, and the exact `MAJOR.MINOR.PATCH` per release.
- Architectures: `linux/amd64`, `linux/arm64`.

## Troubleshooting

- **`config not found at /etc/wireguard/wg0.conf`** — the config isn't mounted at
  the path implied by `BIFROST_INTERFACE`. Mount it read-only there.
- **Container is `unhealthy`** — no recent handshake. Check `docker logs bifrost`
  for the data path and recovery attempts; verify the server is reachable and
  (for DDNS) that the hostname resolves.
- **Dependent app has no network** — it must use `network_mode: "service:bifrost"`
  and publish its ports on the **bifrost** container, not on itself.
- **Recovery seems slow** — recovery starts only after `BIFROST_STALE_AFTER`; lower
  it (and the backoffs) for faster reaction, raise it to be conservative.

## Development

```bash
# unit tests (bats) via a throwaway container
docker run --rm -v "$PWD:/code" -w /code alpine:3.20 sh -c \
  'apk add --no-cache bats bash >/dev/null && chmod +x tests/mocks/wg src/*.sh && bats tests/*.bats'

# lint
docker run --rm -v "$PWD:/code" -w /code koalaman/shellcheck:stable src/*.sh tests/e2e/run.sh
docker run --rm -i hadolint/hadolint < Dockerfile

# end-to-end tunnel test (needs a Linux Docker host; ~minutes)
docker build -t bifrost:ci . && ./tests/e2e/run.sh
```

Commits follow [Conventional Commits](https://www.conventionalcommits.org/);
releases and the GHCR image are cut automatically by release-please.

## License

MIT (see [`LICENSE`](LICENSE)). The built image aggregates GPL-2.0 binaries
(`wireguard-tools`, `wireguard-go`); see [`NOTICE`](NOTICE).
