ARG ALPINE_VERSION=3.14
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
    org.opencontainers.image.title="DNS over TLS upstream server" \
    org.opencontainers.image.description="Runs a local DNS server connected to Cloudflare DNS server 1.1.1.1 over TLS (and more)"
EXPOSE 53/udp
ENV \
    PROVIDERS=cloudflare \
    PRIVATE_ADDRESS=127.0.0.1/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,169.254.0.0/16,::1/128,fc00::/7,fe80::/10,::ffff:7f00:1/104,::ffff:a00:0/104,::ffff:a9fe:0/112,::ffff:ac10:0/108,::ffff:c0a8:0/112 \
    LISTENINGPORT=53 \
    VERBOSITY=1 \
    VERBOSITY_DETAILS=0 \
    VALIDATION_LOGLEVEL=0 \
    CACHING=on \
    IPV4=on \
    IPV6=off \
    BLOCK_MALICIOUS=on \
    BLOCK_SURVEILLANCE=off \
    BLOCK_ADS=off \
    BLOCK_IPS= \
    BLOCK_HOSTNAMES= \
    UNBLOCK= \
    CHECK_DNS=on \
    UPDATE_PERIOD=24h
ENTRYPOINT /entrypoint
HEALTHCHECK --interval=5m --timeout=15s --start-period=5s --retries=1 CMD /entrypoint healthcheck
WORKDIR /unbound
RUN apk --update --no-cache add unbound libcap ca-certificates && \
    mv /usr/sbin/unbound . && \
    mv /etc/ssl/certs/ca-certificates.crt . && \
    chown 1000 -R . && \
    chmod 700 . && \
    chmod 400 ca-certificates.crt && \
    chmod 500 unbound && \
    setcap 'cap_net_bind_service=+ep' unbound && \
    apk del libcap && \
    rm -rf /var/cache/apk/* /etc/unbound/* /usr/sbin/unbound-*
COPY --from=build --chown=1000 /tmp/gobuild/entrypoint /entrypoint
USER 1000
# Downloads and install some files
RUN /entrypoint build
