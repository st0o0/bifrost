load test_helper

setup() {
  ROOT="$(bifrost_repo_root)"
  setup_path_with_mocks
  TMP="$(mktemp -d)"
  cat > "$TMP/wg0.conf" <<'EOF'
[Interface]
PrivateKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
Address = 10.100.99.2/32

[Peer]
PublicKey = BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=
Endpoint = example.com:51820
AllowedIPs = 10.100.99.1/32
EOF
}

teardown() {
  rm -rf "$TMP"
}

@test "reresolve script exists and is bash" {
  head -n1 "$ROOT/src/reresolve-dns.sh" | grep -q 'bash'
}

@test "reresolve runs without error against a config (mock wg)" {
  run bash "$ROOT/src/reresolve-dns.sh" "$TMP/wg0.conf"
  [ "$status" -eq 0 ]
}
