ARG ALPINE_VERSION=3.11
ARG GO_VERSION=1.14

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder
RUN apk --update add git
WORKDIR /tmp/gobuild
ENV CGO_ENABLED=0
COPY go.mod go.sum ./
RUN go mod download 2>&1
COPY internal/ ./internal/
COPY cmd/main.go .
RUN go test ./...
RUN go build -ldflags="-s -w" -o entrypoint main.go

FROM alpine:${ALPINE_VERSION}
ARG BUILD_DATE
ARG VCS_REF
ARG VERSION
ENV VERSION=$VERSION \
    BUILD_DATE=$BUILD_DATE \
    VCS_REF=$VCS_REF
LABEL \
    org.opencontainers.image.authors="quentin.mcgaw@gmail.com" \
    org.opencontainers.image.created=$BUILD_DATE \
    org.opencontainers.image.version=$VERSION \
    org.opencontainers.image.revision=$VCS_REF \
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
    BLOCK_MALICIOUS=on \
    BLOCK_SURVEILLANCE=off \
    BLOCK_ADS=off \
    BLOCK_IPS= \
    BLOCK_HOSTNAMES= \
    UNBLOCK=
ENTRYPOINT /entrypoint
HEALTHCHECK --interval=5m --timeout=15s --start-period=5s --retries=1 CMD /entrypoint healthcheck
WORKDIR /unbound
RUN adduser nonrootuser -D -H --uid 1000 && \
    echo "http://dl-cdn.alpinelinux.org/alpine/edge/main" > /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
    apk --update --progress -q add ca-certificates unbound libcap && \
    mv /usr/sbin/unbound . && \
    mv /etc/ssl/certs/ca-certificates.crt . && \
    touch include.conf && \
    chown nonrootuser -R . && \
    chmod 700 . && \
    chmod 400 ca-certificates.crt include.conf && \
    chmod 500 unbound && \
    setcap 'cap_net_bind_service=+ep' unbound && \
    apk del libcap && \
    rm -rf /var/cache/apk/* /etc/unbound/* /usr/sbin/unbound-*
COPY --from=builder --chown=nonrootuser /tmp/gobuild/entrypoint /entrypoint
USER nonrootuser
