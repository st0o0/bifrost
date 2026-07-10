#!/usr/bin/env bash
# End-to-end: real WireGuard tunnel between two bifrost containers, a deterministic
# DDNS-style server IP change, and recovery via each stage.
#   Scenario A (resolve): BIFROST_RESOLVE=on recovers the IP change in place.
#   Scenario B (reconnect): BIFROST_RESOLVE=off, BIFROST_RECONNECT=on recovers via wg-quick down/up.
# Configs are injected via named volumes (cross-platform). On Windows Git Bash run:
#   MSYS_NO_PATHCONV=1 bash tests/e2e/run.sh
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
cleanup

wgimg() { docker run --rm -i --entrypoint wg "$IMAGE" "$@"; }
put_conf() { docker volume create "$1" >/dev/null; docker run --rm -i --entrypoint sh -v "$1:/etc/wireguard" "$IMAGE" -c 'cat > /etc/wireguard/wg0.conf'; }

s_priv="$(wgimg genkey)"; s_pub="$(printf '%s' "$s_priv" | wgimg pubkey)"
c_priv="$(wgimg genkey)"; c_pub="$(printf '%s' "$c_priv" | wgimg pubkey)"

put_conf "$SVOL" <<EOF
[Interface]
Address = 10.13.13.1/24
ListenPort = 51820
PrivateKey = $s_priv

[Peer]
PublicKey = $c_pub
AllowedIPs = 10.13.13.2/32
EOF

put_conf "$CVOL" <<EOF
[Interface]
Address = 10.13.13.2/24
PrivateKey = $c_priv

[Peer]
PublicKey = $s_pub
Endpoint = bifrost-e2e-server:51820
AllowedIPs = 10.13.13.1/32
PersistentKeepalive = 5
EOF

docker network create --subnet "$SUBNET" "$NET" >/dev/null

start_server() { # $1 ip
  docker run -d --rm --name bifrost-e2e-server --network "$NET" --ip "$1" \
    --cap-add NET_ADMIN --device /dev/net/tun \
    -e BIFROST_RESOLVE=off -e BIFROST_RECONNECT=off -e BIFROST_HEALTHCHECK=off \
    -v "$SVOL:/etc/wireguard:ro" "$IMAGE" >/dev/null
}
client_handshake() {
  docker exec bifrost-e2e-client wg show wg0 latest-handshakes 2>/dev/null \
    | awk 'BEGIN{m=0}{if($2+0>m)m=$2}END{print m+0}'
}

run_scenario() {  # $1=label  $2..=extra client env flags
  local label="$1"; shift
  echo "== [$label] starting server@$SERVER_IP1 and client =="
  start_server "$SERVER_IP1"
  # shellcheck disable=SC2086
  docker run -d --rm --name bifrost-e2e-client --network "$NET" --ip "$CLIENT_IP" \
    --cap-add NET_ADMIN --device /dev/net/tun \
    -e BIFROST_CHECK_INTERVAL=3 -e BIFROST_STALE_AFTER=15 \
    -e BIFROST_RESOLVE_BACKOFF=2 -e BIFROST_RECONNECT_BACKOFF=2 \
    -e BIFROST_HEALTHCHECK=off "$@" \
    -v "$CVOL:/etc/wireguard:ro" "$IMAGE" >/dev/null

  local ok=0 hs
  for _ in $(seq 1 30); do hs="$(client_handshake)"; [ "$hs" -gt 0 ] && { ok=1; break; }; sleep 2; done
  [ "$ok" -eq 1 ] || { echo "[$label] FAIL: no initial handshake"; docker logs bifrost-e2e-server; docker logs bifrost-e2e-client; return 1; }
  local before="$hs"
  echo "[$label] initial handshake $before"

  echo "== [$label] moving server to $SERVER_IP2 =="
  docker rm -f bifrost-e2e-server >/dev/null
  start_server "$SERVER_IP2"

  # Budget generously: resolve recovery can wait on Docker DNS re-propagation
  # after the server container-name swap (observed ~2 min on Docker Desktop).
  ok=0
  for _ in $(seq 1 80); do hs="$(client_handshake)"; [ "$hs" -gt "$before" ] && { ok=1; break; }; sleep 3; done
  [ "$ok" -eq 1 ] || { echo "[$label] FAIL: no recovery"; docker logs bifrost-e2e-client; return 1; }
  echo "[$label] OK: recovered (new handshake $hs > $before)"

  docker rm -f bifrost-e2e-client bifrost-e2e-server >/dev/null 2>&1 || true
}

run_probe_scenario() {
  echo "== [probe] starting server + client (probe on, STALE_AFTER=999) =="
  start_server "$SERVER_IP1"
  docker run -d --rm --name bifrost-e2e-client --network "$NET" --ip "$CLIENT_IP" \
    --cap-add NET_ADMIN --device /dev/net/tun \
    -e BIFROST_STALE_AFTER=999 \
    -e BIFROST_PROBE=on -e BIFROST_PROBE_INTERVAL=2 -e BIFROST_PROBE_FAILS=2 -e BIFROST_PROBE_TIMEOUT=1 \
    -e BIFROST_RESOLVE=off -e BIFROST_RECONNECT=on -e BIFROST_RECONNECT_BACKOFF=2 -e BIFROST_RECONNECT_RETRIES=10 \
    -e BIFROST_HEALTHCHECK=off \
    -v "$CVOL:/etc/wireguard:ro" "$IMAGE" >/dev/null

  local ok=0 hs
  for _ in $(seq 1 30); do hs="$(client_handshake)"; [ "$hs" -gt 0 ] && { ok=1; break; }; sleep 2; done
  [ "$ok" -eq 1 ] || { echo "[probe] FAIL: no initial handshake"; docker logs bifrost-e2e-server; docker logs bifrost-e2e-client; return 1; }
  local before="$hs"
  echo "[probe] initial handshake $before"

  echo "== [probe] killing server; only the probe can trigger (STALE_AFTER=999) =="
  docker rm -f bifrost-e2e-server >/dev/null
  ok=0
  for _ in $(seq 1 20); do
    docker logs bifrost-e2e-client 2>&1 | grep -q "probe — all targets down" && { ok=1; break; }
    sleep 2
  done
  [ "$ok" -eq 1 ] || { echo "[probe] FAIL: probe did not trigger recovery"; docker logs bifrost-e2e-client; return 1; }
  echo "[probe] OK: probe triggered recovery fast (well before STALE_AFTER=999)"

  echo "== [probe] bringing server back -> recovers =="
  start_server "$SERVER_IP1"
  ok=0
  for _ in $(seq 1 40); do hs="$(client_handshake)"; [ "$hs" -gt "$before" ] && { ok=1; break; }; sleep 3; done
  [ "$ok" -eq 1 ] || { echo "[probe] FAIL: no recovery after server returned"; docker logs bifrost-e2e-client; return 1; }
  echo "[probe] OK: recovered after server returned"
  docker rm -f bifrost-e2e-client bifrost-e2e-server >/dev/null 2>&1 || true
}

run_scenario "resolve"   -e BIFROST_RESOLVE=on  -e BIFROST_RECONNECT=off
run_scenario "reconnect" -e BIFROST_RESOLVE=off -e BIFROST_RECONNECT=on
run_probe_scenario
echo "E2E PASSED (resolve + reconnect + probe)"
