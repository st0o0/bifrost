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
docker logs -f bifrost           # shows data path + reresolve loop
docker inspect --format '{{.State.Health.Status}}' bifrost
```

Publish the ports of any tunneled service on the **bifrost** container, and give
the service `network_mode: "service:bifrost"`.

## Configuration

| Variable | Default | Purpose |
|---|---|---|
| `BIFROST_INTERFACE` | `wg0` | Config file name (`/etc/wireguard/<name>.conf`) |
| `BIFROST_RERESOLVE_INTERVAL` | `30` | Seconds between DDNS re-resolve runs (`0` disables) |
| `BIFROST_HEALTH_MAX_HANDSHAKE_AGE` | `180` | Healthcheck threshold in seconds |

Requires `cap_add: NET_ADMIN`. `/dev/net/tun` is only needed for the userspace
fallback (older kernels / hosts without the WireGuard module).

## License

MIT (see `LICENSE`). Vendors the official WireGuard `reresolve-dns.sh`
(GPL-2.0); see `NOTICE`.
