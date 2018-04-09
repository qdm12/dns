FROM alpine:3.7
LABEL maintainer="quentin.mcgaw@gmail.com" \
      description="Runs a DNS server connected to the secured Cloudflare DNS server 1.1.1.1" \
      download="?MB" \
      size="9.58MB" \
      ram="6MB" \
      cpu_usage="Very Low" \
      github="https://github.com/qdm12/cloudflare-dns-server"
RUN apk add --update --no-cache -q --progress unbound ca-certificates && \
    rm -r /etc/unbound/unbound.conf /var/cache/apk/*
COPY unbound.conf /etc/unbound/unbound.conf
EXPOSE 53
ENTRYPOINT unbound -d