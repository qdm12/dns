FROM alpine:3.8 AS downloader
RUN apk add -q --progress wget perl-xml-xpath && \
    wget -q https://www.internic.net/domain/named.root -O named.root && \
    echo "602f28581292bf5e50c8137c955173e6  named.root" > hashes.md5 && \
    md5sum -c hashes.md5 && \
    wget -q https://data.iana.org/root-anchors/root-anchors.xml -O root-anchors.xml && \
    echo "1b2a628d1ff22d4dc7645cfc89f21b6a575526439c6706ecf853e6fff7099dc8  root-anchors.xml" > hashes.sha256 && \
    sha256sum -c hashes.sha256 && \
    KEYTAGS=$(xpath -q -e '/TrustAnchor/KeyDigest/KeyTag/node()' root-anchors.xml) && \
    ALGORITHMS=$(xpath -q -e '/TrustAnchor/KeyDigest/Algorithm/node()' root-anchors.xml) && \
    DIGESTTYPES=$(xpath -q -e '/TrustAnchor/KeyDigest/DigestType/node()' root-anchors.xml) && \
    DIGESTS=$(xpath -q -e '/TrustAnchor/KeyDigest/Digest/node()' root-anchors.xml) && \
    i=1 && \
    while [ 1 ]; do \
      KEYTAG=$(echo $KEYTAGS | cut -d" " -f$i); \
      [ "$KEYTAG" != "" ] || break; \
      ALGORITHM=$(echo $ALGORITHMS | cut -d" " -f$i); \
      DIGESTTYPE=$(echo $DIGESTTYPES | cut -d" " -f$i); \
      DIGEST=$(echo $DIGESTS | cut -d" " -f$i); \
      echo ". IN DS $KEYTAG $ALGORITHM $DIGESTTYPE $DIGEST" >> /root.key; \
      i=`expr $i + 1`; \
    done;

FROM alpine:3.8
LABEL maintainer="quentin.mcgaw@gmail.com" \
      description="Runs a DNS server connected to the secured Cloudflare DNS server 1.1.1.1" \
      download="5MB" \
      size="12.2MB" \
      ram="6MB" \
      cpu_usage="Very Low" \
      github="https://github.com/qdm12/cloudflare-dns-server"
EXPOSE 53/udp
ENV VERBOSITY=1 \
    VERBOSITY_DETAILS=2
HEALTHCHECK --interval=10m --timeout=4s --start-period=3s --retries=2 CMD wget -qO- duckduckgo.com &> /dev/null || exit 1
COPY --from=downloader /named.root /etc/unbound/root.hints
COPY --from=downloader /root.key /etc/unbound/root.key
RUN apk add --update --no-cache -q --progress unbound && \
    rm -rf /var/cache/apk/* /etc/unbound/unbound.conf && \
    echo "#Add Unbound configuration below" > /etc/unbound/include.conf && \
    chown unbound /etc/unbound/root.key
COPY unbound.conf blocks-malicious.conf entrypoint.sh /etc/unbound/
ENTRYPOINT /etc/unbound/entrypoint.sh
