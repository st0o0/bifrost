# Contributing

Thanks for your interest in bifrost!

## Development

Everything runs in Docker — no local toolchain required.

```bash
# unit tests (bats)
docker run --rm -v "$PWD:/code" -w /code alpine:3.20 sh -c \
  'apk add --no-cache bats bash >/dev/null && chmod +x tests/mocks/wg src/*.sh && bats tests/*.bats'

# lint
docker run --rm -v "$PWD:/code" -w /code koalaman/shellcheck:stable src/*.sh tests/e2e/run.sh
docker run --rm -i hadolint/hadolint < Dockerfile

# build + end-to-end tunnel test (needs a Linux Docker host; ~minutes)
docker build -t bifrost:ci . && ./tests/e2e/run.sh
```

## Pull requests

- Branch from `main` and open a PR — CI (lint, unit, build, e2e) runs on pull
  requests.
- Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/)
  (`feat:`, `fix:`, `docs:`, `ci:`, `chore:`, …); commitlint enforces this and
  release-please derives the version and changelog from them.
- Keep changes focused and add or update tests for any behavior change.
- Shell scripts are POSIX `sh` (except the bats tests) and must pass
  `shellcheck`; the `Dockerfile` must pass `hadolint`.

## Reporting bugs / requesting features

Use the issue templates. For security issues see [`SECURITY.md`](SECURITY.md).

By contributing you agree to the [Code of Conduct](CODE_OF_CONDUCT.md).
