ARG BASE_IMAGE=alpine
ARG ALPINE_VERSION=3.10

FROM ${BASE_IMAGE}:${ALPINE_VERSION} AS updated
WORKDIR /tmp/updated
RUN wget -q https://raw.githubusercontent.com/qdm12/updated/master/files/named.root.updated -O root.hints && \
    wget -q https://raw.githubusercontent.com/qdm12/updated/master/files/root.key.updated -O root.key
WORKDIR /tmp/updated/work
RUN wget -q https://raw.githubusercontent.com/qdm12/updated/master/files/malicious-hostnames.updated -O malicious-hostnames && \
    wget -q https://raw.githubusercontent.com/qdm12/updated/master/files/nsa-hostnames.updated -O nsa-hostnames && \
    wget -q https://raw.githubusercontent.com/qdm12/updated/master/files/malicious-ips.updated -O malicious-ips && \
    while read hostname; do echo "local-zone: \""$hostname"\" static" >> blocks-malicious.conf; done < malicious-hostnames && \
    while read ip; do echo "private-address: $ip" >> blocks-malicious.conf; done < malicious-ips && \
    tar -cjf /tmp/updated/blocks-malicious.bz2 blocks-malicious.conf && \
    while read hostname; do echo "local-zone: \""$hostname"\" static" >> blocks-nsa.conf; done < nsa-hostnames && \
    tar -cjf /tmp/updated/blocks-nsa.bz2 blocks-nsa.conf && \
    rm -rf /tmp/updated/work/*

FROM ${BASE_IMAGE}:${ALPINE_VERSION}
ARG BUILD_DATE
ARG VCS_REF
LABEL \
    org.opencontainers.image.authors="quentin.mcgaw@gmail.com" \
    org.opencontainers.image.created=$BUILD_DATE \
    org.opencontainers.image.version="" \
    org.opencontainers.image.revision=$VCS_REF \
    org.opencontainers.image.url="https://github.com/qdm12/cloudflare-dns-server" \
    org.opencontainers.image.documentation="https://github.com/qdm12/cloudflare-dns-server/blob/master/README.md" \
    org.opencontainers.image.source="https://github.com/qdm12/cloudflare-dns-server" \
    org.opencontainers.image.title="cloudflare-dns-server" \
    org.opencontainers.image.description="Runs a local DNS server connected to Cloudflare DNS server 1.1.1.1 over TLS (and more)" \
    image-size="28MB" \
    ram-usage="13.2MB to 70MB" \
    cpu-usage="Low"
EXPOSE 53/udp
ENV VERBOSITY=1 \
    VERBOSITY_DETAILS=0 \
    BLOCK_MALICIOUS=on \
    BLOCK_NSA=off \
    UNBLOCK= \
    LISTENINGPORT=53 \
    PROVIDERS=cloudflare \
    CACHING=on
ENTRYPOINT /unbound/entrypoint.sh
HEALTHCHECK --interval=5m --timeout=15s --start-period=5s --retries=1 \
    CMD LISTENINGPORT=${LISTENINGPORT:-53}; dig @127.0.0.1 +short +time=1 duckduckgo.com -p $LISTENINGPORT &> /dev/null; [ $? = 0 ] || exit 1
WORKDIR /unbound
RUN apk --update --progress -q add ca-certificates unbound bind-tools libcap && \
    adduser nonrootuser -D -H --uid 1000 && \
    chown nonrootuser /usr/sbin/unbound && \
    chmod 500 /usr/sbin/unbound && \
    setcap 'cap_net_bind_service=+ep' /usr/sbin/unbound && \
    apk del libcap && \
    mv /etc/ssl/certs/ca-certificates.crt . && \
    rm -rf /var/cache/apk/* /etc/unbound /usr/sbin/unbound-*
COPY --from=updated /tmp/updated/root.hints .
COPY --from=updated /tmp/updated/root.key .
COPY --from=updated /tmp/updated/blocks-malicious.bz2 .
COPY --from=updated /tmp/updated/blocks-nsa.bz2 .
COPY unbound.conf entrypoint.sh ./
RUN chown nonrootuser -R . && \
    chmod 700 . && \
    chmod 600 unbound.conf && \
    chmod 700 entrypoint.sh && \
    chmod 400 root.hints root.key ca-certificates.crt *.bz2 && \
    mv /usr/sbin/unbound .
USER nonrootuser