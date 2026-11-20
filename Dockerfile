# syntax=docker/dockerfile:1@sha256:572cebbc2b15e6f84c077742952674889cb63cd5e9b80abdffcbe587b6f19183

FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280 AS build
ARG TARGETARCH
ARG VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /bifrost ./cmd/bifrost

FROM scratch
LABEL org.opencontainers.image.title="bifrost" \
      org.opencontainers.image.description="Minimal native-Go WireGuard client that keeps dependents tunneled across DDNS endpoint IP changes without a netns teardown" \
      org.opencontainers.image.source="https://github.com/st0o0/bifrost" \
      org.opencontainers.image.documentation="https://github.com/st0o0/bifrost#readme" \
      org.opencontainers.image.licenses="MIT"
COPY --from=build /bifrost /bifrost
COPY LICENSE NOTICE /

HEALTHCHECK --interval=30s --timeout=10s --start-period=45s --retries=3 \
  CMD ["/bifrost", "healthcheck"]

ENTRYPOINT ["/bifrost"]
