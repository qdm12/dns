ARG BASE_IMAGE=alpine
ARG ALPINE_VERSION=3.10

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
    image-size="26.9MB" \
    ram-usage="13.2MB to 70MB" \
    cpu-usage="Low"
EXPOSE 53/udp
ENV VERBOSITY=1 \
    VERBOSITY_DETAILS=0 \
    BLOCK_MALICIOUS=on \
    BLOCK_NSA=off \
    UNBLOCK= \
    LISTENINGPORT=53 \
    PROVIDER=cloudflare
ENTRYPOINT /etc/unbound/entrypoint.sh
HEALTHCHECK --interval=5m --timeout=15s --start-period=5s --retries=1 \
    CMD LISTENINGPORT=${LISTENINGPORT:-53}; dig @127.0.0.1 +short +time=1 duckduckgo.com -p $LISTENINGPORT &> /dev/null; [ $? = 0 ] || exit 1
RUN apk --update --no-cache --progress -q add ca-certificates unbound bind-tools libcap && \
    setcap 'cap_net_bind_service=+ep' /usr/sbin/unbound && \
    apk del libcap && \
    rm -rf /var/cache/apk/* /etc/unbound/* /usr/sbin/unbound-anchor /usr/sbin/unbound-checkconf /usr/sbin/unbound-control /usr/sbin/unbound-control-setup /usr/sbin/unbound-host && \
    adduser nonrootuser -D -H --uid 1000 && \
    wget -q https://raw.githubusercontent.com/qdm12/updated/master/files/named.root.updated -O /etc/unbound/root.hints && \
    wget -q https://raw.githubusercontent.com/qdm12/updated/master/files/root.key.updated -O /etc/unbound/root.key && \
    cd /tmp && \
    wget -q https://raw.githubusercontent.com/qdm12/updated/master/files/malicious-hostnames.updated -O malicious-hostnames && \
    wget -q https://raw.githubusercontent.com/qdm12/updated/master/files/nsa-hostnames.updated -O nsa-hostnames && \
    wget -q https://raw.githubusercontent.com/qdm12/updated/master/files/malicious-ips.updated -O malicious-ips && \
    while read hostname; do echo "local-zone: \""$hostname"\" static" >> blocks-malicious.conf; done < malicious-hostnames && \
    while read ip; do echo "private-address: $ip" >> blocks-malicious.conf; done < malicious-ips && \
    tar -cjf /etc/unbound/blocks-malicious.bz2 blocks-malicious.conf && \
    while read hostname; do echo "local-zone: \""$hostname"\" static" >> blocks-nsa.conf; done < nsa-hostnames && \
    tar -cjf /etc/unbound/blocks-nsa.bz2 blocks-nsa.conf && \
    rm -f /tmp/*
COPY unbound.conf entrypoint.sh /etc/unbound/
RUN chown nonrootuser -R /etc/unbound && \
    chmod 700 /etc/unbound && \
    chmod 600 /etc/unbound/unbound.conf && \
    chmod 500 /etc/unbound/entrypoint.sh && \
    chmod 400 /etc/unbound/root.hints /etc/unbound/root.key /etc/unbound/*.bz2
USER nonrootuser
