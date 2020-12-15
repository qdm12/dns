ARG ALPINE_VERSION=3.12
ARG GO_VERSION=1.15

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder
RUN apk --update add git
ENV CGO_ENABLED=0
ARG GOLANGCI_LINT_VERSION=v1.33.0
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s ${GOLANGCI_LINT_VERSION}
WORKDIR /tmp/gobuild
COPY .golangci.yml .
COPY go.mod go.sum ./
RUN go mod download 2>&1
COPY cmd/main.go cmd/app/main.go
COPY internal/ ./internal/
RUN go test ./...
RUN golangci-lint run --timeout=10m
ARG VERSION=unknown
ARG BUILD_DATE="an unknown date"
ARG COMMIT=unknown
RUN go build -o entrypoint -trimpath -ldflags="-s -w \
    -X 'main.version=$VERSION' \
    -X 'main.buildDate=$BUILD_DATE' \
    -X 'main.commit=$COMMIT'" \
    cmd/app/main.go

FROM alpine:${ALPINE_VERSION}
ARG VERSION=unknown
ARG BUILD_DATE="an unknown date"
ARG COMMIT=unknown
LABEL \
    org.opencontainers.image.authors="quentin.mcgaw@gmail.com" \
    org.opencontainers.image.version=$VERSION \
    org.opencontainers.image.created=$BUILD_DATE \
    org.opencontainers.image.revision=$COMMIT \
    org.opencontainers.image.url="https://github.com/qdm12/cloudflare-dns-server" \
    org.opencontainers.image.documentation="https://github.com/qdm12/cloudflare-dns-server/blob/master/README.md" \
    org.opencontainers.image.source="https://github.com/qdm12/cloudflare-dns-server" \
    org.opencontainers.image.title="DNS over TLS upstream server" \
    org.opencontainers.image.description="Runs a local DNS server connected to Cloudflare DNS server 1.1.1.1 over TLS (and more)"
EXPOSE 53/udp
ENV \
    PROVIDERS=cloudflare \
    PRIVATE_ADDRESS=127.0.0.1/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,169.254.0.0/16,::1/128,fc00::/7,fe80::/10,::ffff:0:0/96 \
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
    CHECK_UNBOUND=on \
    UPDATE_PERIOD=24h
ENTRYPOINT /entrypoint
HEALTHCHECK --interval=5m --timeout=15s --start-period=5s --retries=1 CMD /entrypoint healthcheck
WORKDIR /unbound
RUN apk --update --no-cache add unbound libcap ca-certificates && \
    mv /usr/sbin/unbound . && \
    mv /etc/ssl/certs/ca-certificates.crt . && \
    touch include.conf && \
    chown 1000 -R . && \
    chmod 700 . && \
    chmod 400 ca-certificates.crt include.conf && \
    chmod 500 unbound && \
    setcap 'cap_net_bind_service=+ep' unbound && \
    apk del libcap && \
    rm -rf /var/cache/apk/* /etc/unbound/* /usr/sbin/unbound-*
COPY --from=builder --chown=1000 /tmp/gobuild/entrypoint /entrypoint
USER 1000
