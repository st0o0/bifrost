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
