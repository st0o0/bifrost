#!/bin/sh
# bifrost shared helpers. POSIX sh. Licensed MIT.

bifrost_interface() {
    printf '%s' "${BIFROST_INTERFACE:-wg0}"
}

bifrost_config_path() {
    printf '%s' "/etc/wireguard/$(bifrost_interface).conf"
}

# Returns 0 if kernel WireGuard is usable, 1 otherwise.
# Tests/operators may force the result with BIFROST_FORCE_DATAPATH.
bifrost_kernel_supported() {
    case "${BIFROST_FORCE_DATAPATH:-}" in
        kernel) return 0 ;;
        userspace) return 1 ;;
    esac
    if [ -d /sys/module/wireguard ]; then
        return 0
    fi
    if ip link add dev bifrost_probe type wireguard 2>/dev/null; then
        ip link del dev bifrost_probe 2>/dev/null || true
        return 0
    fi
    return 1
}

# Parse an on/off-style boolean. Returns 0 (true) or 1 (false/unknown).
bifrost_bool() {
    case "$(printf '%s' "${1:-}" | tr '[:upper:]' '[:lower:]')" in
        on|true|yes|1) return 0 ;;
        *) return 1 ;;
    esac
}

# Exit 1 with a clear message unless VALUE is a recognized on/off token.
bifrost_require_bool() {  # $1=name $2=value
    case "$(printf '%s' "${2:-}" | tr '[:upper:]' '[:lower:]')" in
        on|true|yes|1|off|false|no|0) : ;;
        *) echo "bifrost: $1 must be on/off, got '${2:-}'" >&2; exit 1 ;;
    esac
}

# Exit 1 with a clear message unless VALUE is a non-negative integer.
bifrost_require_int() {  # $1=name $2=value
    case "${2:-}" in
        ''|*[!0-9]*) echo "bifrost: $1 must be a non-negative integer, got '${2:-}'" >&2; exit 1 ;;
    esac
}

# Newest peer handshake timestamp (epoch seconds; 0 if none).
bifrost_handshake_ts() {  # $1=iface
    wg show "$1" latest-handshakes 2>/dev/null \
      | awk 'BEGIN{m=0}{if($2+0>m)m=$2}END{print m+0}'
}

# Age in seconds of the newest handshake (999999 if none yet).
bifrost_handshake_age() {  # $1=iface
    _ts="$(bifrost_handshake_ts "$1")"
    if [ "${_ts:-0}" -le 0 ]; then
        echo 999999
    else
        echo $(( $(date +%s) - _ts ))
    fi
}

# Exponential backoff seconds for 1-based attempt N with base B, capped at 60.
bifrost_backoff() {  # $1=attempt $2=base
    _n="$1"; _d="$2"; _i=1
    while [ "$_i" -lt "$_n" ]; do _d=$(( _d * 2 )); _i=$(( _i + 1 )); done
    [ "$_d" -gt 60 ] && _d=60
    echo "$_d"
}
