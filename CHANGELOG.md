# Changelog

## [0.2.2](https://github.com/st0o0/bifrost/compare/v0.2.1...v0.2.2) (2026-09-12)


### Features

* decouple release-please from build workflow ([5caf19e](https://github.com/st0o0/bifrost/commit/5caf19ea2ae86149caf27c01f7cace5ec03bd024))


### Refactoring

* migrate to shared reusable workflows ([f953582](https://github.com/st0o0/bifrost/commit/f953582390bf34508be15a3dd1983fae8ff5b5f4))
* rename CI jobs for cleaner GitHub check names ([ec4a9d8](https://github.com/st0o0/bifrost/commit/ec4a9d890a9b510d02f7dff1803aab7760e8e321))


### Dependencies

* bump golang from 1.27rc2-alpine to 1.27-alpine ([cce3d01](https://github.com/st0o0/bifrost/commit/cce3d014356da7d38e2924c0ce9c5c123c34849b))
* bump golang.org/x/crypto v0.54.0 -&gt; v0.57.0, add .trivyignore ([88bed7c](https://github.com/st0o0/bifrost/commit/88bed7c038daf36a363ad634927a7f9b0dc4c154))
* bump hadolint/hadolint-action in the actions-all group ([4a18a85](https://github.com/st0o0/bifrost/commit/4a18a85aff58211c2b3a94f57459835956225107))

## [0.2.1](https://github.com/st0o0/bifrost/compare/v0.2.0...v0.2.1) (2026-08-16)


### Features

* **config:** add BIFROST_METRICS and BIFROST_METRICS_ADDR settings ([992085d](https://github.com/st0o0/bifrost/commit/992085d8bd5b5041b6dc31cdec33d54fcdb0f635))
* **metrics:** add Prometheus collector, stats counters, and HTTP server ([b6e0ff8](https://github.com/st0o0/bifrost/commit/b6e0ff8e064e1db2b9ab47ac7d4228d7b6f650fe))
* **metrics:** expand Prometheus metrics with device, interface, recovery, and uptime ([d2cb5ec](https://github.com/st0o0/bifrost/commit/d2cb5ecf87a3764f7eddcda9155ae31aef3a239b))
* migrate from stdlib log to log/slog structured logging ([46a3dd3](https://github.com/st0o0/bifrost/commit/46a3dd3ebe17e262a955a6591f5b8f44fc94f341))
* **probe:** return ProbeResult with RTT from AllDown ([208c9fc](https://github.com/st0o0/bifrost/commit/208c9fc843e22273bd1a182da0f78357d3a7f89c))
* **recovery:** add OnResolve and OnReconnect callbacks ([8749c74](https://github.com/st0o0/bifrost/commit/8749c74a3b54556791c488551887d0d2ae901a89))
* **wg:** detect endpoint IP changes during resolve ([00d461f](https://github.com/st0o0/bifrost/commit/00d461f9ecb3aea98e23aca54f1a03b2af11c8a9))
* wire Prometheus metrics into supervisor and main ([51f6757](https://github.com/st0o0/bifrost/commit/51f6757d3fe0c047a223547eb9618b564044cb99))

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
