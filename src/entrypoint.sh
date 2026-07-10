#!/bin/sh
# bifrost entrypoint (PID 1). Licensed MIT.
set -eu
DIR="$(cd "$(dirname "$0")" && pwd)"
. "$DIR/lib.sh"

IFACE="$(bifrost_interface)"
CONF="$(bifrost_config_path)"
INTERVAL="${BIFROST_RERESOLVE_INTERVAL:-30}"

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

loop_pid=""
if [ "$INTERVAL" -gt 0 ]; then
    (
        while :; do
            sleep "$INTERVAL"
            "$DIR/reresolve-dns.sh" "$CONF" || true
        done
    ) &
    loop_pid=$!
    echo "bifrost: reresolve loop running every ${INTERVAL}s (pid $loop_pid)"
else
    echo "bifrost: reresolve disabled (BIFROST_RERESOLVE_INTERVAL=0)"
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
    while :; do
        sleep 3600 &
        wait $!
    done
fi
