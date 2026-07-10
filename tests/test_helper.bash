# Shared bats setup helpers.
bifrost_repo_root() {
  cd "$BATS_TEST_DIRNAME/.." && pwd
}

setup_path_with_mocks() {
  PATH="$BATS_TEST_DIRNAME/mocks:$PATH"
  export PATH
}
