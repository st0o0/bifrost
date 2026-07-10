#!/bin/sh
# bifrost resolver: re-resolve each peer's hostname endpoint and apply via `wg set`.
# Licensed MIT.
set -eu
DIR="$(cd "$(dirname "$0")" && pwd)"
. "$DIR/lib.sh"

IFACE="$(bifrost_interface)"
CONF="${1:-$(bifrost_config_path)}"

_pubkey=""
_endpoint=""

_flush() {
    [ -n "$_pubkey" ] && [ -n "$_endpoint" ] || { _pubkey=""; _endpoint=""; return 0; }
    _host="${_endpoint%:*}"
    # Skip bare IPv4 literals (nothing to re-resolve); anything with a
    # non-digit/non-dot character is treated as a hostname (or bracketed IPv6).
    case "$_host" in
        *[!0-9.]*) wg set "$IFACE" peer "$_pubkey" endpoint "$_endpoint" 2>/dev/null || true ;;
    esac
    _pubkey=""; _endpoint=""
}

while IFS= read -r line || [ -n "$line" ]; do
    _l="${line%%#*}"
    _key="$(printf '%s' "${_l%%=*}" | tr -d '[:space:]')"
    _val="$(printf '%s' "${_l#*=}" | tr -d '[:space:]')"
    case "$_l" in
        \[*\]*) _flush ;;
    esac
    case "$_key" in
        PublicKey) _pubkey="$_val" ;;
        Endpoint)  _endpoint="$_val" ;;
    esac
done < "$CONF"
_flush
