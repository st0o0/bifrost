# syntax=docker/dockerfile:1

FROM scratch
LABEL org.opencontainers.image.title="bifrost" \
      org.opencontainers.image.description="Minimal native-Go WireGuard client that keeps dependents tunneled across DDNS endpoint IP changes without a netns teardown" \
      org.opencontainers.image.source="https://github.com/st0o0/bifrost" \
      org.opencontainers.image.documentation="https://github.com/st0o0/bifrost#readme" \
      org.opencontainers.image.licenses="MIT"
COPY bifrost /bifrost
COPY LICENSE NOTICE /

HEALTHCHECK --interval=30s --timeout=10s --start-period=45s --retries=3 \
  CMD ["/bifrost", "healthcheck"]

ENTRYPOINT ["/bifrost"]
