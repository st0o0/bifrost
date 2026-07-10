load test_helper

setup() {
  ROOT="$(bifrost_repo_root)"
  setup_path_with_mocks
  TMP="$(mktemp -d)"
  export WG_MOCK_LOG="$TMP/wgset.log"
  : > "$WG_MOCK_LOG"
  cat > "$TMP/wg0.conf" <<'EOF'
[Interface]
Address = 10.13.13.2/32
PrivateKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=

[Peer]
PublicKey = HOSTKEYAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
Endpoint = vpn.example.com:51820
AllowedIPs = 10.13.13.1/32

[Peer]
PublicKey = IPKEYBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=
Endpoint = 203.0.113.9:51820
AllowedIPs = 10.13.13.3/32
EOF
}
teardown() { rm -rf "$TMP"; }

@test "resolver sets the hostname peer's endpoint" {
  BIFROST_INTERFACE=wg0 run "$ROOT/src/resolve.sh" "$TMP/wg0.conf"
  [ "$status" -eq 0 ]
  grep -q "set wg0 peer HOSTKEYAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= endpoint vpn.example.com:51820" "$WG_MOCK_LOG"
}
@test "resolver skips the bare-IPv4 peer" {
  BIFROST_INTERFACE=wg0 run "$ROOT/src/resolve.sh" "$TMP/wg0.conf"
  ! grep -q "IPKEYBBB" "$WG_MOCK_LOG"
}
