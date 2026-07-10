#!/bin/sh
# bifrost entrypoint (PID 1) — supervisor with escalating recovery. MIT.
set -eu
DIR="$(cd "$(dirname "$0")" && pwd)"
. "$DIR/lib.sh"

IFACE="$(bifrost_interface)"
CONF="$(bifrost_config_path)"

CHECK_INTERVAL="${BIFROST_CHECK_INTERVAL:-30}"
STALE_AFTER="${BIFROST_STALE_AFTER:-135}"
RESOLVE_RETRIES="${BIFROST_RESOLVE_RETRIES:-5}"
RESOLVE_BACKOFF="${BIFROST_RESOLVE_BACKOFF:-5}"
RECONNECT_RETRIES="${BIFROST_RECONNECT_RETRIES:-5}"
RECONNECT_BACKOFF="${BIFROST_RECONNECT_BACKOFF:-5}"

# Fail fast on bad configuration.
bifrost_require_int  BIFROST_CHECK_INTERVAL     "$CHECK_INTERVAL"
bifrost_require_int  BIFROST_STALE_AFTER        "$STALE_AFTER"
bifrost_require_int  BIFROST_RESOLVE_RETRIES    "$RESOLVE_RETRIES"
bifrost_require_int  BIFROST_RESOLVE_BACKOFF    "$RESOLVE_BACKOFF"
bifrost_require_int  BIFROST_RECONNECT_RETRIES  "$RECONNECT_RETRIES"
bifrost_require_int  BIFROST_RECONNECT_BACKOFF  "$RECONNECT_BACKOFF"
bifrost_require_bool BIFROST_RESOLVE            "${BIFROST_RESOLVE:-on}"
bifrost_require_bool BIFROST_RECONNECT          "${BIFROST_RECONNECT:-on}"
bifrost_require_bool BIFROST_HEALTHCHECK        "${BIFROST_HEALTHCHECK:-on}"
bifrost_require_int  BIFROST_HEALTH_STALE_AFTER "${BIFROST_HEALTH_STALE_AFTER:-180}"

[ "$CHECK_INTERVAL" -ge 1 ] || { echo "bifrost: BIFROST_CHECK_INTERVAL must be >= 1" >&2; exit 1; }

if [ ! -f "$CONF" ]; then
    echo "bifrost: config not found at $CONF — mount your WireGuard .conf there" >&2
    exit 1
fi

if bifrost_kernel_supported; then
    echo "bifrost: using kernel WireGuard data path"
else
    echo "bifrost: kernel module unavailable — using userspace (wireguard-go)"
    export WG_QUICK_USERSPACE_IMPLEMENTATION=wireguard-go
fi

echo "bifrost: bringing up $IFACE"
wg-quick up "$CONF"

recover() {
    _baseline="$(bifrost_handshake_ts "$IFACE")"
    if bifrost_bool "${BIFROST_RESOLVE:-on}"; then
        _a=1
        while [ "$_a" -le "$RESOLVE_RETRIES" ]; do
            echo "bifrost: resolve attempt $_a/$RESOLVE_RETRIES"
            "$DIR/resolve.sh" "$CONF" || true
            sleep "$(bifrost_backoff "$_a" "$RESOLVE_BACKOFF")"
            [ "$(bifrost_handshake_ts "$IFACE")" -gt "$_baseline" ] && { echo "bifrost: recovered via resolve"; return 0; }
            _a=$(( _a + 1 ))
        done
    fi
    if bifrost_bool "${BIFROST_RECONNECT:-on}"; then
        _a=1
        while [ "$_a" -le "$RECONNECT_RETRIES" ]; do
            echo "bifrost: reconnect attempt $_a/$RECONNECT_RETRIES (wg-quick down/up)"
            wg-quick down "$CONF" 2>/dev/null || true
            wg-quick up "$CONF" 2>/dev/null || true
            sleep "$(bifrost_backoff "$_a" "$RECONNECT_BACKOFF")"
            [ "$(bifrost_handshake_ts "$IFACE")" -gt "$_baseline" ] && { echo "bifrost: recovered via reconnect"; return 0; }
            _a=$(( _a + 1 ))
        done
    fi
    echo "bifrost: recovery exhausted (resolve+reconnect); will retry" >&2
    return 1
}

supervise() {
    while :; do
        sleep "$CHECK_INTERVAL"
        _age="$(bifrost_handshake_age "$IFACE")"
        if [ "$_age" -gt "$STALE_AFTER" ]; then
            echo "bifrost: handshake stale (${_age}s > ${STALE_AFTER}s) — recovery"
            recover || true
        fi
    done
}

loop_pid=""
if bifrost_bool "${BIFROST_RESOLVE:-on}" || bifrost_bool "${BIFROST_RECONNECT:-on}"; then
    supervise &
    loop_pid=$!
    echo "bifrost: supervisor running (check ${CHECK_INTERVAL}s, stale>${STALE_AFTER}s)"
else
    echo "bifrost: recovery disabled (resolve and reconnect both off)"
fi

term() {
    echo "bifrost: received signal, stopping"
    [ -n "$loop_pid" ] && kill "$loop_pid" 2>/dev/null || true
    exit 0
}
trap term TERM INT

if [ -n "$loop_pid" ]; then
    wait "$loop_pid"
else
    while :; do sleep 3600 & wait $!; done
fi
