load test_helper

setup() {
  ROOT="$(bifrost_repo_root)"
  setup_path_with_mocks
  . "$ROOT/src/lib.sh"
  TMP="$(mktemp -d)"
  cat > "$TMP/wg0.conf" <<'EOF'
[Interface]
Address = 10.13.13.2/32
PrivateKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=

[Peer]
PublicKey = BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=
Endpoint = vpn.example.com:51820
AllowedIPs = 10.13.13.1/32, 10.0.30.0/24, 10.50.0.10/32, 0.0.0.0/0
EOF
}
teardown() { rm -rf "$TMP"; }

@test "probe targets keep /32 hosts, drop ranges and 0.0.0.0/0" {
  run bifrost_probe_targets "$TMP/wg0.conf"
  [ "$status" -eq 0 ]
  echo "$output" | grep -qx "10.13.13.1"
  echo "$output" | grep -qx "10.50.0.10"
  ! echo "$output" | grep -q "10.0.30.0"
  ! echo "$output" | grep -q "0.0.0.0"
}

@test "BIFROST_PROBE_HOST overrides AllowedIPs derivation" {
  export BIFROST_PROBE_HOST="9.9.9.9, 8.8.8.8"
  run bifrost_probe_targets "$TMP/wg0.conf"
  echo "$output" | grep -qx "9.9.9.9"
  echo "$output" | grep -qx "8.8.8.8"
  ! echo "$output" | grep -q "10.13.13.1"
}

@test "probe_once is up when any target responds" {
  export PING_MOCK_UP="10.50.0.10"
  run bifrost_probe_once 2 <<EOF
10.13.13.1
10.50.0.10
EOF
  [ "$status" -eq 0 ]
}

@test "probe_once is down when all targets fail" {
  export PING_MOCK_UP=""
  run bifrost_probe_once 2 <<EOF
10.13.13.1
10.50.0.10
EOF
  [ "$status" -eq 1 ]
}
