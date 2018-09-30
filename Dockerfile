FROM alpine:3.8
LABEL maintainer="quentin.mcgaw@gmail.com" \
      description="Runs a DNS server connected to the secured Cloudflare DNS server 1.1.1.1" \
      download="5MB" \
      size="12.2MB" \
      ram="6MB" \
      cpu_usage="Very Low" \
      github="https://github.com/qdm12/cloudflare-dns-server"
EXPOSE 53/udp
RUN apk add --update --no-cache -q --progress unbound && \
    rm -rf /etc/unbound/unbound.conf /etc/unbound/root.hints /var/cache/apk/* && \
    touch /etc/unbound/include.conf && \
    wget http://www.internic.net/domain/named.root -O /etc/unbound/root.hints && \
    echo ". IN DS 19036 8 2 49AAC11D7B6F6446702E54A1607371607A1A41855200FD2CE1CDDE32F24E8FB5" > /etc/unbound/root.key && \
    echo ". IN DS 20326 8 2 E06D44B80B8F1D39A95C0B0D7C65D08458E880409BBC683457104237C7F8EC8D" >> /etc/unbound/root.key && \
    chown unbound /etc/unbound/root.key
HEALTHCHECK --interval=10m --timeout=4s --start-period=3s --retries=1 CMD wget -qO- duckduckgo.com &> /dev/null || exit 1
ENV VERBOSITY=1
ENTRYPOINT sed -i "s/verbosity: 2/verbosity: $VERBOSITY/g" /etc/unbound/unbound.conf && \
           $(grep -Fq 127.0.0.1 /etc/resolv.conf) || echo "WARNING: The domain name is not set to 127.0.0.1 so the healthcheck will likely be irrelevant!" && \
           unbound -d $1
COPY unbound.conf blocks-malicious.conf /etc/unbound/
