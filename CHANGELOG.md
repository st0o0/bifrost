# Changelog

## 0.1.0 (2026-07-10)


### Features

* add healthcheck toggle and rename threshold to BIFROST_HEALTH_STALE_AFTER ([70642f5](https://github.com/st0o0/bifrost/commit/70642f5ec3751a0eb6cfbf73fdb55e5bec6ab386))
* add lib helpers (bool/int validation, handshake age, backoff) ([29d8104](https://github.com/st0o0/bifrost/commit/29d8104db3da5e0c0dc81eb12ee5d072cc4ab7fe))
* bifrost — WireGuard client image with in-place DDNS reconnect ([1faa199](https://github.com/st0o0/bifrost/commit/1faa1995c49f2f2dccd24331749800dcc2fb37f6))
* own MIT resolver, drop vendored GPL reresolve-dns ([746747b](https://github.com/st0o0/bifrost/commit/746747bd473f323fe6c19208999a6de4d386a2f9))
* supervisor loop with escalating resolve/reconnect recovery ([6893125](https://github.com/st0o0/bifrost/commit/689312546fed5e9c14a4a1e93ae8b0f510569691))


### Bug Fixes

* **ci:** satisfy shellcheck (quote SC2295, disable SC1091 for dynamic source) ([8e2e404](https://github.com/st0o0/bifrost/commit/8e2e404d611d25b505e69bf26b80815f2e402281))
* reconcile Dockerfile/CI/README after dropping reresolve-dns.sh ([104ea7d](https://github.com/st0o0/bifrost/commit/104ea7d432d3f377fbb656f75091b890c20808eb))
* validate healthcheck env vars and reject zero check interval ([be1b85b](https://github.com/st0o0/bifrost/commit/be1b85baa38bb0d580e87597e8d94d05fb3ade96))
