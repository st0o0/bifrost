# Contributing

Thanks for your interest in bifrost!

## Development

bifrost is a Go module. The logic packages test anywhere; the Linux-only
packages (`internal/wg`, the ICMP pinger, `main`) build on Linux.

```bash
# tests + vet (Go)
go test ./...
go vet ./...

# lint
golangci-lint run          # golangci-lint v2
docker run --rm -i hadolint/hadolint < Dockerfile

# build + end-to-end tunnel test (needs a Linux Docker host; ~minutes)
docker build -t bifrost:ci . && ./tests/e2e/run.sh
```

## Pull requests

- Branch from `main` and open a PR — CI (test, lint, build, e2e) runs on pull
  requests.
- Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/)
  (`feat:`, `fix:`, `docs:`, `ci:`, `chore:`, …); commitlint enforces this and
  release-please derives the version and changelog from them.
- Keep changes focused and add or update tests for any behavior change.
- Go code must pass `go vet` and `golangci-lint`; the `Dockerfile` must pass
  `hadolint`.

## Reporting bugs / requesting features

Use the issue templates. For security issues see [`SECURITY.md`](SECURITY.md).

By contributing you agree to the [Code of Conduct](CODE_OF_CONDUCT.md).
