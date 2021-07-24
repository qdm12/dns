ARG BUILDPLATFORM=linux/amd64

ARG ALPINE_VERSION=3.13
ARG GO_VERSION=1.16

ARG GOLANGCI_LINT_VERSION=v1.41.1
ARG XCPUTRANSLATE_VERSION=v0.6.0

FROM --platform=${BUILDPLATFORM} qmcgaw/binpot:golangci-lint-${GOLANGCI_LINT_VERSION} AS golangci-lint
FROM --platform=${BUILDPLATFORM} qmcgaw/xcputranslate:${XCPUTRANSLATE_VERSION} AS xcputranslate

FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS base
RUN apk --update --no-cache add git g++
ENV CGO_ENABLED=0
WORKDIR /tmp/gobuild
COPY --from=golangci-lint /bin /go/bin/golangci-lint
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY pkg/ ./pkg/
COPY internal/ ./internal/

FROM --platform=${BUILDPLATFORM} base AS test
# Note on the go race detector:
# - we set CGO_ENABLED=1 to have it enabled
# - we installed g++ to support the race detector
ENV CGO_ENABLED=1
ENTRYPOINT go test -race -coverprofile=coverage.txt ./...

FROM --platform=${BUILDPLATFORM} base AS lint
COPY .golangci.yml ./
RUN golangci-lint run --timeout=10m

FROM --platform=$BUILDPLATFORM base AS tidy
RUN git init && \
    git config user.email ci@localhost && \
    git config user.name ci && \
    git add -A && git commit -m ci && \
    sed -i '/\/\/ indirect/d' go.mod && \
    go mod tidy && \
    git diff --exit-code -- go.mod

FROM --platform=${BUILDPLATFORM} base AS build
COPY --from=xcputranslate /xcputranslate /usr/local/bin/xcputranslate
ARG TARGETPLATFORM
ARG VERSION=unknown
ARG BUILD_DATE="an unknown date"
ARG COMMIT=unknown
RUN GOARCH="$(xcputranslate translate -field arch -targetplatform ${TARGETPLATFORM})" \
    GOARM="$(xcputranslate translate -field arm -targetplatform ${TARGETPLATFORM})" \
    go build -trimpath -ldflags="-s -w \
    -X 'main.version=$VERSION' \
    -X 'main.buildDate=$BUILD_DATE' \
    -X 'main.commit=$COMMIT' \
    " -o entrypoint cmd/main.go
RUN apk --update --no-cache add libcap && \
    setcap 'cap_net_bind_service=+ep' entrypoint && \
    apk del libcap

FROM --platform=${BUILDPLATFORM} alpine:${ALPINE_VERSION} AS alpine
RUN apk --update add ca-certificates

FROM scratch
COPY --from=alpine --chown=1000 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 53/udp
ENTRYPOINT ["/entrypoint"]
HEALTHCHECK --interval=5m --timeout=15s --start-period=5s --retries=1 CMD ["/entrypoint", "healthcheck"]
USER 1000
ENV \
    UPSTREAM_TYPE=DoT \
    DOT_RESOLVERS=cloudflare,google \
    DOH_RESOLVERS=cloudflare,google \
    DNS_PLAINTEXT_RESOLVERS=cloudflare \
    DOT_TIMEOUT=3s \
    DOH_TIMEOUT=3s \
    LISTENING_PORT=53 \
    LOG_LEVEL=info \
    CACHE_TYPE=lru \
    IPV4=on \
    IPV6=off \
    BLOCK_MALICIOUS=on \
    BLOCK_SURVEILLANCE=off \
    BLOCK_ADS=off \
    BLOCK_IPS= \
    BLOCK_CIDRS= \
    BLOCK_HOSTNAMES= \
    ALLOWED_HOSTNAMES= \
    CHECK_DNS=on \
    UPDATE_PERIOD=24h
ARG VERSION=unknown
ARG BUILD_DATE="an unknown date"
ARG COMMIT=unknown
LABEL \
    org.opencontainers.image.authors="quentin.mcgaw@gmail.com" \
    org.opencontainers.image.version=$VERSION \
    org.opencontainers.image.created=$BUILD_DATE \
    org.opencontainers.image.revision=$COMMIT \
    org.opencontainers.image.url="https://github.com/qdm12/dns" \
    org.opencontainers.image.documentation="https://github.com/qdm12/dns/blob/master/README.md" \
    org.opencontainers.image.source="https://github.com/qdm12/dns" \
    org.opencontainers.image.title="DNS over TLS or HTTPS upstream server" \
    org.opencontainers.image.description="Runs a local DNS server connected to nameservers with DNS over TLS or DNS over HTTPs"
COPY --from=build --chown=1000 /tmp/gobuild/entrypoint /entrypoint

# Downloads and install some files
# TODO once DNSSEC is operational
# RUN /entrypoint build
