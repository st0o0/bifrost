#!/usr/bin/env bash
# End-to-end test: a real WireGuard tunnel between two bifrost containers, then a
# deterministic DDNS-style endpoint IP change, proving the reresolve loop recovers
# the tunnel in place without a container/netns restart.
#
# NOTE: the vendored reresolve-dns.sh only re-resolves a peer whose last handshake
# is older than ~135s (upstream behaviour). Recovery after the IP change therefore
# takes ~135-150s; the whole run takes ~2.5-5 min. That is expected, not a hang.
#
# Configs are injected via named volumes (not host bind-mounts) so the script runs
# identically on Linux CI and on Docker Desktop. On Windows Git Bash, invoke it as
#   MSYS_NO_PATHCONV=1 bash tests/e2e/run.sh
# so the ":/etc/wireguard" and "/dev/net/tun" arguments are not path-mangled.
set -euo pipefail

IMAGE="bifrost:ci"
NET="bifrost-e2e-net"
SUBNET="172.28.0.0/24"
SERVER_IP1="172.28.0.10"
SERVER_IP2="172.28.0.11"
CLIENT_IP="172.28.0.20"
SVOL="bifrost-e2e-scfg"
CVOL="bifrost-e2e-ccfg"

cleanup() {
  docker rm -f bifrost-e2e-client bifrost-e2e-server >/dev/null 2>&1 || true
  docker network rm "$NET" >/dev/null 2>&1 || true
  docker volume rm "$SVOL" "$CVOL" >/dev/null 2>&1 || true
}
trap cleanup EXIT
cleanup   # clear any leftovers from a previous run

# Run the image's `wg` tool (override ENTRYPOINT; -i so `wg pubkey` reads stdin).
wgimg() { docker run --rm -i --entrypoint wg "$IMAGE" "$@"; }

# Write a wg0.conf (read from stdin) into a fresh named volume.
put_conf() {  # $1 = volume name
  docker volume create "$1" >/dev/null
  docker run --rm -i --entrypoint sh -v "$1:/etc/wireguard" "$IMAGE" \
    -c 'cat > /etc/wireguard/wg0.conf'
}

s_priv="$(wgimg genkey)"; s_pub="$(printf '%s' "$s_priv" | wgimg pubkey)"
c_priv="$(wgimg genkey)"; c_pub="$(printf '%s' "$c_priv" | wgimg pubkey)"

put_conf "$SVOL" <<EOF
[Interface]
Address = 10.100.99.1/24
ListenPort = 51820
PrivateKey = $s_priv

[Peer]
PublicKey = $c_pub
AllowedIPs = 10.100.99.2/32
EOF

# Endpoint is the server's container name -> Docker DNS. A replacement server with
# the same name but a different --ip is a faithful DDNS IP change.
put_conf "$CVOL" <<EOF
[Interface]
Address = 10.100.99.2/24
PrivateKey = $c_priv

[Peer]
PublicKey = $s_pub
Endpoint = bifrost-e2e-server:51820
AllowedIPs = 10.100.99.1/32
PersistentKeepalive = 5
EOF

docker network create --subnet "$SUBNET" "$NET" >/dev/null

start_server() {  # $1 = static IP
  docker run -d --rm --name bifrost-e2e-server \
    --network "$NET" --ip "$1" \
    --cap-add NET_ADMIN --device /dev/net/tun \
    -e BIFROST_RERESOLVE_INTERVAL=0 \
    -v "$SVOL:/etc/wireguard:ro" \
    "$IMAGE" >/dev/null
}

client_handshake() {  # newest handshake epoch (0 if none yet)
  docker exec bifrost-e2e-client wg show wg0 latest-handshakes 2>/dev/null \
    | awk 'BEGIN{m=0}{if($2+0>m)m=$2}END{print m+0}'
}

echo "== starting server at $SERVER_IP1 and client =="
start_server "$SERVER_IP1"
docker run -d --rm --name bifrost-e2e-client \
  --network "$NET" --ip "$CLIENT_IP" \
  --cap-add NET_ADMIN --device /dev/net/tun \
  -e BIFROST_RERESOLVE_INTERVAL=5 -e BIFROST_HEALTH_MAX_HANDSHAKE_AGE=30 \
  -v "$CVOL:/etc/wireguard:ro" \
  "$IMAGE" >/dev/null

echo "== waiting for initial handshake =="
ok=0
for _ in $(seq 1 30); do
  hs="$(client_handshake)"
  [ "$hs" -gt 0 ] && { ok=1; break; }
  sleep 2
done
[ "$ok" -eq 1 ] || { echo "FAIL: no initial handshake"; echo "-- server --"; docker logs bifrost-e2e-server; echo "-- client --"; docker logs bifrost-e2e-client; exit 1; }
before="$hs"
echo "OK: initial handshake (epoch $before)"

echo "== replacing server with a new IP $SERVER_IP2 (simulated DDNS change) =="
docker rm -f bifrost-e2e-server >/dev/null
start_server "$SERVER_IP2"

echo "== waiting for reresolve-driven recovery (upstream ~135s threshold) =="
ok=0
for _ in $(seq 1 60); do            # up to ~300s
  hs="$(client_handshake)"
  [ "$hs" -gt "$before" ] && { ok=1; break; }
  sleep 5
done
[ "$ok" -eq 1 ] || { echo "FAIL: no new handshake after IP change"; docker logs bifrost-e2e-client; exit 1; }
echo "OK: reresolve recovered the tunnel (new handshake $hs > $before)"
echo "E2E PASSED"
