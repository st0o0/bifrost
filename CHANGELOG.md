# Changelog

## 0.1.0 (2026-07-10)


### Features

* add healthcheck toggle and rename threshold to BIFROST_HEALTH_STALE_AFTER ([423dfc8](https://github.com/st0o0/bifrost/commit/423dfc85e7bf2e56287d0286aecc2ae3205720a3))
* add lib helpers (bool/int validation, handshake age, backoff) ([8b4ea29](https://github.com/st0o0/bifrost/commit/8b4ea29a599e68d15f1bf226ae4ae8bf76239d21))
* bifrost — WireGuard client image with in-place DDNS reconnect ([b13f3e2](https://github.com/st0o0/bifrost/commit/b13f3e2d5db5d74b69b1f40b3fd42777042fb004))
* own MIT resolver, drop vendored GPL reresolve-dns ([818de42](https://github.com/st0o0/bifrost/commit/818de42bb38fb45a5376538e04c5893a2ff93f87))
* supervisor loop with escalating resolve/reconnect recovery ([b64f885](https://github.com/st0o0/bifrost/commit/b64f8852d09f63283ff27cd2c9c84b0516f0322d))


### Bug Fixes

* **ci:** satisfy shellcheck (quote SC2295, disable SC1091 for dynamic source) ([27b5efc](https://github.com/st0o0/bifrost/commit/27b5efc789212a011812370facc53f40f18f4547))
* reconcile Dockerfile/CI/README after dropping reresolve-dns.sh ([2ba8d08](https://github.com/st0o0/bifrost/commit/2ba8d08c5ef5c93841a34c21adc898ea784adbb7))
* validate healthcheck env vars and reject zero check interval ([a3df4e6](https://github.com/st0o0/bifrost/commit/a3df4e6bdb906d1a082463100cd3dbafdb4f7ed9))
