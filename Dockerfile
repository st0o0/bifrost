# syntax=docker/dockerfile:1@sha256:ecfaec9ed6d810b56388c508f4121597bfbba70d41a6dfeee4d8cad5f295fc32

FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:4cb7ac979db5fcc41cae44b2227ba5ab8a51e8807f40d9ba4dee20a0ad960b5b AS build
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
