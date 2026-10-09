# syntax=docker/dockerfile:1@sha256:4edf897a3ffa55b89f906fc8cc78afdb3f1834cc9c7083565e611a8a7d5fe99e

FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:738d1cf061836894ff6bb8c33881080ac66de8cf0586615012a0c8f592649cfa AS build
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
