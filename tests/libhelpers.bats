load test_helper

setup() {
  ROOT="$(bifrost_repo_root)"
  setup_path_with_mocks
  . "$ROOT/src/lib.sh"
}

@test "bifrost_bool true tokens" {
  for v in on On ON true yes 1; do run bifrost_bool "$v"; [ "$status" -eq 0 ] || { echo "failed for $v"; return 1; }; done
}
@test "bifrost_bool false/unknown tokens" {
  for v in off false no 0 "" garbage; do run bifrost_bool "$v"; [ "$status" -eq 1 ] || { echo "failed for $v"; return 1; }; done
}
@test "bifrost_require_int accepts an integer" {
  run bifrost_require_int NAME 42
  [ "$status" -eq 0 ]
}
@test "bifrost_require_int rejects non-numeric" {
  run bifrost_require_int BIFROST_CHECK_INTERVAL abc
  [ "$status" -eq 1 ]
  [[ "$output" == *"BIFROST_CHECK_INTERVAL"* ]]
}
@test "bifrost_require_bool rejects garbage" {
  run bifrost_require_bool BIFROST_RESOLVE maybe
  [ "$status" -eq 1 ]
}
@test "bifrost_backoff is exponential capped at 60" {
  run bifrost_backoff 1 5; [ "$output" = "5" ]
  run bifrost_backoff 2 5; [ "$output" = "10" ]
  run bifrost_backoff 3 5; [ "$output" = "20" ]
  run bifrost_backoff 5 5; [ "$output" = "60" ]   # 80 -> capped
}
@test "bifrost_handshake_ts reads newest" {
  export WG_MOCK_HANDSHAKES="$(printf 'A\t100\nB\t250')"
  run bifrost_handshake_ts wg0
  [ "$output" = "250" ]
}
@test "bifrost_handshake_age is large when no handshake" {
  export WG_MOCK_HANDSHAKES="$(printf 'A\t0')"
  run bifrost_handshake_age wg0
  [ "$output" -ge 999999 ]
}
