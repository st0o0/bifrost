load test_helper

setup() {
  ROOT="$(bifrost_repo_root)"
}

@test "bifrost_interface defaults to wg0" {
  . "$ROOT/src/lib.sh"
  run bifrost_interface
  [ "$status" -eq 0 ]
  [ "$output" = "wg0" ]
}

@test "bifrost_interface honors BIFROST_INTERFACE" {
  export BIFROST_INTERFACE=wg1
  . "$ROOT/src/lib.sh"
  run bifrost_interface
  [ "$output" = "wg1" ]
}

@test "bifrost_config_path builds the conf path" {
  export BIFROST_INTERFACE=wg3
  . "$ROOT/src/lib.sh"
  run bifrost_config_path
  [ "$output" = "/etc/wireguard/wg3.conf" ]
}

@test "bifrost_kernel_supported honors force=kernel" {
  export BIFROST_FORCE_DATAPATH=kernel
  . "$ROOT/src/lib.sh"
  run bifrost_kernel_supported
  [ "$status" -eq 0 ]
}

@test "bifrost_kernel_supported honors force=userspace" {
  export BIFROST_FORCE_DATAPATH=userspace
  . "$ROOT/src/lib.sh"
  run bifrost_kernel_supported
  [ "$status" -eq 1 ]
}
