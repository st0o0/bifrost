#!/bin/sh
# bifrost healthcheck: fail unless a peer handshaked recently. Licensed MIT.
set -eu
DIR="$(cd "$(dirname "$0")" && pwd)"
. "$DIR/lib.sh"

if ! bifrost_bool "${BIFROST_HEALTHCHECK:-on}"; then
    echo "bifrost: healthcheck disabled"
    exit 0
fi

IFACE="$(bifrost_interface)"
MAX_AGE="${BIFROST_HEALTH_STALE_AFTER:-180}"

now="$(date +%s)"
newest=0

# `wg show <iface> latest-handshakes` prints: <pubkey><TAB><epoch-seconds>
TAB="$(printf '\t')"
handshakes="$(wg show "$IFACE" latest-handshakes 2>/dev/null || true)"
IFS='
'
for line in $handshakes; do
    ts="${line#*"$TAB"}"
    case "$ts" in
        ''|*[!0-9]*) continue ;;
    esac
    if [ "$ts" -gt "$newest" ]; then
        newest="$ts"
    fi
done
unset IFS

if [ "$newest" -eq 0 ]; then
    echo "bifrost: no handshake yet on $IFACE" >&2
    exit 1
fi

age=$(( now - newest ))
if [ "$age" -gt "$MAX_AGE" ]; then
    echo "bifrost: last handshake ${age}s ago (> ${MAX_AGE}s) on $IFACE" >&2
    exit 1
fi

echo "bifrost: healthy, last handshake ${age}s ago on $IFACE"
exit 0
