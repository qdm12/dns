ARG ALPINE_VERSION=3.13
ARG GO_VERSION=1.16

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS base
RUN apk --update add git
ENV CGO_ENABLED=0
WORKDIR /tmp/gobuild
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY pkg/ ./pkg/
COPY internal/ ./internal/

FROM --platform=$BUILDPLATFORM base AS test
# Note on the go race detector:
# - we set CGO_ENABLED=1 to have it enabled
# - we install g++ to support the race detector
ENV CGO_ENABLED=1
RUN apk --update --no-cache add g++

FROM --platform=$BUILDPLATFORM base AS lint
ARG GOLANGCI_LINT_VERSION=v1.41.1
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
    sh -s -- -b /usr/local/bin ${GOLANGCI_LINT_VERSION}
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

FROM --platform=$BUILDPLATFORM base AS build
COPY --from=qmcgaw/xcputranslate:v0.4.0 /xcputranslate /usr/local/bin/xcputranslate
ARG TARGETPLATFORM
ARG VERSION=unknown
ARG BUILD_DATE="an unknown date"
ARG COMMIT=unknown
RUN GOARCH="$(xcputranslate -field arch -targetplatform ${TARGETPLATFORM})" \
    GOARM="$(xcputranslate -field arm -targetplatform ${TARGETPLATFORM})" \
    go build -trimpath -ldflags="-s -w \
    -X 'main.version=$VERSION' \
    -X 'main.buildDate=$BUILD_DATE' \
    -X 'main.commit=$COMMIT' \
    " -o entrypoint cmd/main.go

FROM alpine:${ALPINE_VERSION}
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
EXPOSE 53/udp
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
    BLOCK_IPNETS= \
    BLOCK_HOSTNAMES= \
    ALLOWED_HOSTNAMES= \
    CHECK_DNS=on \
    UPDATE_PERIOD=24h
ENTRYPOINT /entrypoint
HEALTHCHECK --interval=5m --timeout=15s --start-period=5s --retries=1 CMD /entrypoint healthcheck
COPY --from=build --chown=1000 /tmp/gobuild/entrypoint /entrypoint
USER 1000
# Downloads and install some files
# TODO once DNSSEC is operational
# RUN /entrypoint build
