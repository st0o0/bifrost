load test_helper

setup() {
  ROOT="$(bifrost_repo_root)"
  setup_path_with_mocks
}

@test "entrypoint exits 1 with clear message when config is missing" {
  export BIFROST_INTERFACE=wgdoesnotexist
  run "$ROOT/src/entrypoint.sh"
  [ "$status" -eq 1 ]
  [[ "$output" == *"config not found"* ]]
}

@test "entrypoint rejects non-numeric BIFROST_CHECK_INTERVAL" {
  export BIFROST_INTERFACE=wgdoesnotexist
  export BIFROST_CHECK_INTERVAL=abc
  run "$ROOT/src/entrypoint.sh"
  [ "$status" -eq 1 ]
  [[ "$output" == *"BIFROST_CHECK_INTERVAL"* ]]
}

@test "entrypoint rejects invalid toggle BIFROST_RESOLVE" {
  export BIFROST_INTERFACE=wgdoesnotexist
  export BIFROST_RESOLVE=maybe
  run "$ROOT/src/entrypoint.sh"
  [ "$status" -eq 1 ]
  [[ "$output" == *"BIFROST_RESOLVE"* ]]
}

@test "entrypoint rejects non-numeric BIFROST_HEALTH_STALE_AFTER" {
  export BIFROST_INTERFACE=wgdoesnotexist
  export BIFROST_HEALTH_STALE_AFTER=180s
  run "$ROOT/src/entrypoint.sh"
  [ "$status" -eq 1 ]
  [[ "$output" == *"BIFROST_HEALTH_STALE_AFTER"* ]]
}

@test "entrypoint rejects invalid BIFROST_HEALTHCHECK toggle" {
  export BIFROST_INTERFACE=wgdoesnotexist
  export BIFROST_HEALTHCHECK=enabled
  run "$ROOT/src/entrypoint.sh"
  [ "$status" -eq 1 ]
  [[ "$output" == *"BIFROST_HEALTHCHECK"* ]]
}

@test "entrypoint rejects BIFROST_CHECK_INTERVAL=0" {
  export BIFROST_INTERFACE=wgdoesnotexist
  export BIFROST_CHECK_INTERVAL=0
  run "$ROOT/src/entrypoint.sh"
  [ "$status" -eq 1 ]
  [[ "$output" == *"BIFROST_CHECK_INTERVAL"* ]]
}
