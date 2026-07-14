# Changelog

## [0.2.0](https://github.com/st0o0/bifrost/compare/v0.1.0...v0.2.0) (2026-07-14)


### Features

* add liveness-probe helpers (target derivation, all-down ping round) ([de5d8cb](https://github.com/st0o0/bifrost/commit/de5d8cb9f391a811ad42797dfe6e8aeb3b70b995))
* **config:** Allow WireGuard config via environment variables ([5905309](https://github.com/st0o0/bifrost/commit/5905309da92eefeeaf6d4e8b5cc6947da5d22d2c))
* **go:** add a version subcommand stamped at build time ([6ce6013](https://github.com/st0o0/bifrost/commit/6ce6013c7bd3ffc3c0af9b141346b462dc875db4))
* **go:** add env settings loading and validation ([90b0614](https://github.com/st0o0/bifrost/commit/90b06149d1ce9a05c25cb516cf4ae4755c49af95))
* **go:** add genkey/pubkey/handshake subcommands ([6ad7029](https://github.com/st0o0/bifrost/commit/6ad7029c606853d8a782d3f5d56590bb8d8c731b))
* **go:** add probe target derivation and fail-state ([6fc3996](https://github.com/st0o0/bifrost/commit/6fc3996d569acb1fa6c7091b0b24b2ec5b9e1341))
* **go:** add recovery backoff/episode and healthcheck logic ([43f7a48](https://github.com/st0o0/bifrost/commit/43f7a484749d1b8c4b675b7d825bd66af46884a5))
* **go:** add WireGuard .conf parser ([7914cee](https://github.com/st0o0/bifrost/commit/7914cee80adcdbec5818ac8efa2423fd852fb6f7))
* **go:** native ICMP pinger and all-down decision ([824857a](https://github.com/st0o0/bifrost/commit/824857ab497feadbf8509d5e7b58a937dc6cd9fc))
* **go:** native wg device lifecycle and wgctrl controller ([5c48d97](https://github.com/st0o0/bifrost/commit/5c48d97050f808543bec57f472fe7c71cb51e5a8))
* **go:** scaffold module and cmd/bifrost ([80bce31](https://github.com/st0o0/bifrost/commit/80bce31cc9fdb2fb1ca370de7bce7fbc6f52df8f))
* **go:** supervisor loop with probe + handshake-age triggers ([9960cc1](https://github.com/st0o0/bifrost/commit/9960cc158c6615178cbf38d473bf15ca06686be3))
* **go:** wire main, supervisor, and healthcheck subcommand ([d756bf1](https://github.com/st0o0/bifrost/commit/d756bf12556b5d5eda67f52434e74941f045388d))
* trigger recovery from the liveness probe (fast unreachability detection) ([fbb7be9](https://github.com/st0o0/bifrost/commit/fbb7be97ecc1e78749ae652f765c296ade2803bd))


### Bug Fixes

* floor probe interval/fails/timeout at &gt;= 1 and tighten probe docs ([7cff4fa](https://github.com/st0o0/bifrost/commit/7cff4fa9599ecc08ea9316d78de87db8f8143db1))
* **go:** bring wg link up before routes, clean up on Bring errors, tolerate leftover link/routes ([13672f8](https://github.com/st0o0/bifrost/commit/13672f8361b56246c815c8332f136e799cb57a50))
* **go:** honor ctx cancellation in recovery, keep hostnames in probe targets, add config validation ([91513c0](https://github.com/st0o0/bifrost/commit/91513c04d6291897f99d2fde13ee69b8748a40dd))
* mark tests/mocks/ping executable ([787609d](https://github.com/st0o0/bifrost/commit/787609dbef7b50d4c3887ea16fac90ef3ae6303a))
* **supervisor:** latch first handshake so recovery never goes dormant ([cd6a3eb](https://github.com/st0o0/bifrost/commit/cd6a3eb3dbb7bf524b3688b6dc2ce1652a5be170))

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
