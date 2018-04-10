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
    rm -r /etc/unbound/unbound.conf /var/cache/apk/*
COPY unbound.conf /etc/unbound/unbound.conf
ENTRYPOINT unbound -d
CMD ["-v"]