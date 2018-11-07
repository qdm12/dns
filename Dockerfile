ARG ALPINE_VERSION=3.8

FROM alpine:${ALPINE_VERSION}
ARG BUILD_DATE
ARG VCS_REF
LABEL org.label-schema.schema-version="1.0.0-rc1" \
      maintainer="quentin.mcgaw@gmail.com" \
      org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/qdm12/cloudflare-dns-server" \
      org.label-schema.url="https://github.com/qdm12/cloudflare-dns-server" \
      org.label-schema.vcs-description="Runs a local DNS server connected to Cloudflare DNS server 1.1.1.1 over TLS" \
      org.label-schema.vcs-usage="https://github.com/qdm12/cloudflare-dns-server/blob/master/README.md#setup" \
      org.label-schema.docker.cmd="docker run -d -p 53:53/udp --dns=127.0.0.1 qmcgaw/cloudflare-dns-server" \
      org.label-schema.docker.cmd.devel="docker run -it --rm -p 53:53/udp --dns=127.0.0.1 -e VERBOSITY=3 -e VERBOSITY_DETAILS=3 -e BLOCK_MALICIOUS=off qmcgaw/cloudflare-dns-server" \
      org.label-schema.docker.params="VERBOSITY=from 0 (no log) to 5 (full debug log) and defaults to 1,VERBOSITY_DETAILS=0 to 4 and defaults to 0 (higher means more details),BLOCK_MALICIOUS='on' or 'off' and defaults to 'on' (note that it consumes about 50MB of additional RAM)" \
      image-size="12.7MB" \
      ram-usage="13.2MB to 70MB" \
      cpu-usage="Low"
EXPOSE 1053/udp
ENV VERBOSITY=1 \
    VERBOSITY_DETAILS=0 \
    BLOCK_MALICIOUS=on
ENTRYPOINT /etc/unbound/entrypoint.sh
HEALTHCHECK --interval=5m --timeout=15s --start-period=5s --retries=2 CMD if [ "$(nslookup duckduckgo.com 2>nul)" = "" ]; then exit 1; fi
RUN apk --update --no-cache --progress -q add unbound && \
    rm -rf /var/cache/apk/* /etc/unbound/unbound.conf && \
    echo "# Add Unbound configuration below" > /etc/unbound/include.conf && \
    addgroup nonrootgroup && \
    adduser nonrootuser -G nonrootgroup -D -H
COPY --from=qmcgaw/dns-trustanchor /root.key /etc/unbound/root.key
COPY --from=qmcgaw/dns-trustanchor /named.root /etc/unbound/root.hints
COPY --from=qmcgaw/malicious-hostnames /malicious-hostnames.bz2 /etc/unbound/malicious-hostnames.bz2
COPY --from=qmcgaw/malicious-ips /malicious-ips.bz2 /etc/unbound/malicious-ips.bz2
COPY unbound.conf entrypoint.sh /etc/unbound/
RUN chown nonrootuser:nonrootgroup -R /etc/unbound && \
    chmod 700 -R /etc/unbound && \
    chmod 500 /etc/unbound/entrypoint.sh && \
    chmod 400 \
        /etc/unbound/root.hints \
        /etc/unbound/root.key \
        /etc/unbound/unbound.conf \
        /etc/unbound/malicious-ips.bz2 \
        /etc/unbound/malicious-hostnames.bz2
USER nonrootuser
