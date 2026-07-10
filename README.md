# bifrost

A minimal WireGuard **client** container that keeps dependent containers
tunneled across DDNS endpoint IP changes — without tearing down its network
namespace. The Norse rainbow bridge between your networks.

## Why

- **linuxserver/wireguard**: needs the kernel module mounted (`SYS_MODULE`,
  `/lib/modules`), s6, PUID/PGID; a container restart to recover drops every
  container routed through it.
- **gluetun**: full-tunnel killswitch and **cannot use a DDNS hostname**
  endpoint (IP only).

bifrost fills the gap: kernel-first with a userspace fallback, split-tunnel
friendly, and it re-resolves DDNS endpoints **in place** (`wg set`, no restart),
so `network_mode: service:bifrost` dependents stay connected.

## Usage

1. Put your WireGuard client config at `./wg/wg0.conf` (your existing
   linuxserver config works unchanged). For DDNS keep
   `PersistentKeepalive = 25` in the `[Peer]` section.
2. Copy `.env.example` to `.env`.
3. `docker compose up -d`

```bash
docker logs -f bifrost           # shows data path + recovery loop
docker inspect --format '{{.State.Health.Status}}' bifrost
```

Publish the ports of any tunneled service on the **bifrost** container, and give
the service `network_mode: "service:bifrost"`.

## Configuration

bifrost recovers a stale tunnel in two escalating stages per episode: first it
re-resolves DDNS endpoints in place (gentle, `wg set`), retrying with
exponential backoff; if that doesn't restore a handshake it falls back to an
active `wg-quick down/up` (the network namespace is preserved, so dependents
stay attached). Each stage's retries/backoff and the trigger thresholds are
configurable; either stage can be turned off.

| Variable | Default | Purpose |
|---|---|---|
| `BIFROST_INTERFACE` | `wg0` | Config file name (`/etc/wireguard/<name>.conf`) |
| `BIFROST_CHECK_INTERVAL` | `30` | Supervisor: how often to check handshake freshness (seconds) |
| `BIFROST_STALE_AFTER` | `135` | Handshake age (seconds) that triggers a recovery episode |
| `BIFROST_RESOLVE` | `on` | Recovery stage 1: re-resolve DDNS endpoints (gentle, in place) |
| `BIFROST_RESOLVE_RETRIES` | `5` | Stage 1 retry count |
| `BIFROST_RESOLVE_BACKOFF` | `5` | Stage 1 backoff (seconds) |
| `BIFROST_RECONNECT` | `on` | Recovery stage 2: active `wg-quick down/up` (netns preserved) |
| `BIFROST_RECONNECT_RETRIES` | `5` | Stage 2 retry count |
| `BIFROST_RECONNECT_BACKOFF` | `5` | Stage 2 backoff (seconds) |
| `BIFROST_HEALTHCHECK` | `on` | Enable/disable the container healthcheck |
| `BIFROST_HEALTH_STALE_AFTER` | `180` | Healthcheck threshold in seconds |

If you change `BIFROST_INTERFACE` away from `wg0`, mount your config at the
matching `/etc/wireguard/<name>.conf` (the example compose mounts `wg0.conf`).

Requires `cap_add: NET_ADMIN`. `/dev/net/tun` is only needed for the userspace
fallback (older kernels / hosts without the WireGuard module).

## License

MIT (see `LICENSE`). The built image aggregates GPL-2.0 binaries
(`wireguard-tools`, `wireguard-go`); see `NOTICE`.
