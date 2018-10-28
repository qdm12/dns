FROM alpine:3.8
LABEL maintainer="quentin.mcgaw@gmail.com" \
      description="Runs a DNS server connected to Cloudflare DNS server 1.1.1.1 over TLS" \
      download="5MB" \
      size="12.1MB" \
      ram="13.2MB or 70MB" \
      cpu_usage="Very low to low" \
      github="https://github.com/qdm12/cloudflare-dns-server"
EXPOSE 53/udp
ENV VERBOSITY=1 \
    VERBOSITY_DETAILS=0 \
    BLOCK_MALICIOUS=on
HEALTHCHECK --interval=5m --timeout=15s --start-period=5s --retries=2 CMD if [[ "$(nslookup duckduckgo.com 2>nul)" == "" ]]; then echo "Can't resolve duckduckgo.com"; exit 1; fi
COPY --from=qmcgaw/dns-trustanchor /named.root /etc/unbound/root.hints
COPY --from=qmcgaw/dns-trustanchor /root.key /etc/unbound/root.key
RUN echo https://dl-3.alpinelinux.org/alpine/v3.8/main > /etc/apk/repositories && \
	apk add --update --no-cache -q --progress unbound && \
    rm -rf /var/cache/apk/* /etc/unbound/unbound.conf && \
    echo "# Add Unbound configuration below" > /etc/unbound/include.conf && \
    chown unbound /etc/unbound/root.key
COPY --from=qmcgaw/malicious-hostnames /malicious-hostnames.bz2 /etc/unbound/malicious-hostnames.bz2
COPY --from=qmcgaw/malicious-ips /malicious-ips.bz2 /etc/unbound/malicious-ips.bz2
COPY unbound.conf entrypoint.sh /etc/unbound/
RUN chmod +x /etc/unbound/entrypoint.sh
ENTRYPOINT /etc/unbound/entrypoint.sh
