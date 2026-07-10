# Security Policy

## Supported versions

The latest released image (`ghcr.io/st0o0/bifrost:latest`) receives security
updates. Older tags are not maintained — pin to a `MAJOR.MINOR` tag and update
regularly.

## Reporting a vulnerability

Please report security issues **privately** via GitHub's
[private vulnerability reporting](https://github.com/st0o0/bifrost/security/advisories/new)
(Security → Advisories → *Report a vulnerability*). Do **not** open a public
issue for security problems.

You can expect an initial response within a few days. Once a fix is available it
is released and the advisory is published.

## Scope

bifrost bundles `wireguard-tools` and a compiled `wireguard-go`. Vulnerabilities
in those upstream projects should be reported to their maintainers; when a fixed
version is available the image is rebuilt against it. The image is scanned
weekly with Trivy and results appear in the repository's Security tab.
