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
