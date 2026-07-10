# syntax=docker/dockerfile:1

# ---- build wireguard-go (userspace fallback; not in Alpine stable) ----
FROM golang:1.23-alpine AS wggo
ENV CGO_ENABLED=0
# wireguard-go's main package is the MODULE ROOT (there is no cmd/wireguard-go),
# so the produced binary is named `wireguard`. The project stopped publishing
# semantic tags after 2020, so pin the current commit pseudo-version.
RUN go install golang.zx2c4.com/wireguard@v0.0.0-20260522210424-ecfc5a8d5446

# ---- runtime ----
FROM alpine:3.20 AS runtime
RUN apk add --no-cache \
      wireguard-tools \
      iptables \
      iproute2 \
      bash \
      openresolv

COPY --from=wggo /go/bin/wireguard /usr/bin/wireguard-go
COPY src/ /opt/bifrost/
COPY LICENSE NOTICE /opt/bifrost/
RUN chmod +x /opt/bifrost/entrypoint.sh \
             /opt/bifrost/healthcheck.sh \
             /opt/bifrost/resolve.sh

HEALTHCHECK --interval=30s --timeout=10s --start-period=20s --retries=3 \
  CMD /opt/bifrost/healthcheck.sh || exit 1

ENTRYPOINT ["/opt/bifrost/entrypoint.sh"]
