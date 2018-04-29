FROM alpine:3.7
LABEL maintainer="quentin.mcgaw@gmail.com" \
      description="Runs a DNS server connected to the secured Cloudflare DNS server 1.1.1.1" \
      download="4.3MB" \
      size="9.58MB" \
      ram="6MB" \
      cpu_usage="Very Low" \
      github="https://github.com/qdm12/cloudflare-dns-server"
EXPOSE 53/udp
RUN apk add --update --no-cache -q --progress unbound && \
    rm -rf /etc/unbound/unbound.conf /var/cache/apk/*
COPY unbound.conf blocks-malicious.conf blocks.conf /etc/unbound/
HEALTHCHECK --interval=10m --timeout=3s --start-period=3s --retries=1 CMD ping -W 1 -w 2 google.com &> /dev/null || exit 1
# The container DNS is changed to localhost for the healthcheck
ENTRYPOINT echo "nameserver 127.0.0.1" > /etc/resolv.conf && \
           echo "options ndots:0" >> /etc/resolv.conf && \
           unbound -d
CMD ["-v"]
