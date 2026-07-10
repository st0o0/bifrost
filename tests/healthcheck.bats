load test_helper

setup() {
  ROOT="$(bifrost_repo_root)"
  setup_path_with_mocks
}

@test "healthy when handshake is fresh" {
  export WG_MOCK_HANDSHAKES="$(printf 'PUBKEYAAA\t%s' "$(date +%s)")"
  run "$ROOT/src/healthcheck.sh"
  [ "$status" -eq 0 ]
}

@test "unhealthy when handshake is stale" {
  export BIFROST_HEALTH_MAX_HANDSHAKE_AGE=180
  export WG_MOCK_HANDSHAKES="$(printf 'PUBKEYAAA\t%s' "$(( $(date +%s) - 9999 ))")"
  run "$ROOT/src/healthcheck.sh"
  [ "$status" -eq 1 ]
}

@test "unhealthy when there is no handshake (ts=0)" {
  export WG_MOCK_HANDSHAKES="$(printf 'PUBKEYAAA\t0')"
  run "$ROOT/src/healthcheck.sh"
  [ "$status" -eq 1 ]
}

@test "unhealthy when wg output is empty" {
  export WG_MOCK_HANDSHAKES=""
  run "$ROOT/src/healthcheck.sh"
  [ "$status" -eq 1 ]
}

@test "healthy picks newest of multiple peers" {
  now="$(date +%s)"
  export WG_MOCK_HANDSHAKES="$(printf 'PEER1\t%s\nPEER2\t%s' "$(( now - 5000 ))" "$now")"
  run "$ROOT/src/healthcheck.sh"
  [ "$status" -eq 0 ]
}
