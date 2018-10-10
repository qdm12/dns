FROM alpine:3.8 AS rootanchors
RUN apk add -q --progress wget perl-xml-xpath
RUN wget -q https://www.internic.net/domain/named.root -O named.root && \
    echo "602f28581292bf5e50c8137c955173e6  named.root" > hashes.md5 && \
    md5sum -c hashes.md5
RUN wget -q https://data.iana.org/root-anchors/root-anchors.xml -O root-anchors.xml && \
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

FROM alpine:3.8 AS blocks
RUN apk add -q --progress wget ca-certificates sed
RUN hostnames=$(wget -qO- https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts | \
    sed '/\(^[ \|\t]*#\)\|\(^[ ]\+\)\|\(^$\)\|\(^[\n\|\r\|\r\n][ \|\t]*$\)\|\(^127.0.0.1\)\|\(^255.255.255.255\)\|\(^::1\)\|\(^fe80\)\|\(^ff00\)\|\(^ff02\)\|\(^0.0.0.0 0.0.0.0\)/d' | \
    sed 's/\([ \|\t]*#.*$\)\|\(\r\)\|\(0.0.0.0 \)//g')$'\n'$( \
    wget -qO- https://raw.githubusercontent.com/CHEF-KOCH/NSABlocklist/master/HOSTS | \
    sed '/\(^[ \|\t]*#\)\|\(^[ ]\+\)\|\(^$\)\|\(^[\n\|\r\|\r\n][ \|\t]*$\)\|\(^127.0.0.1\)/d' | \
    sed 's/\([ \|\t]*#.*$\)\|\(\r\)\|\(0.0.0.0 \)//g')$'\n'$( \
    wget -qO- https://raw.githubusercontent.com/k0nsl/unbound-blocklist/master/blocks.conf | \
    sed '/\(^[ \|\t]*#\)\|\(^[ ]\+\)\|\(^$\)\|\(^[\n\|\r\|\r\n][ \|\t]*$\)\|\(^local-data\)/d' | \
    sed 's/\([ \|\t]*#.*$\)\|\(\r\)\|\(local-zone: \"\)\|\(\" redirect\)//g')$'\n'$( \
    wget -qO- https://raw.githubusercontent.com/notracking/hosts-blocklists/master/domains.txt | \
    sed '/\(^[ \|\t]*#\)\|\(^[ ]\+\)\|\(^$\)\|\(^[\n\|\r\|\r\n][ \|\t]*$\)\|\(::$\)/d' | \
    sed 's/\([ \|\t]*#.*$\)\|\(\r\)\|\(address=\/\)\|\(\/0.0.0.0$\)//g')$'\n'$( \
    wget -qO- https://raw.githubusercontent.com/notracking/hosts-blocklists/master/hostnames.txt | \
    sed '/\(^[ \|\t]*#\)\|\(^[ ]\+\)\|\(^$\)\|\(^[\n\|\r\|\r\n][ \|\t]*$\)\|\(^::\)/d' | \
    sed 's/\([ \|\t]*#.*$\)\|\(\r\)\|\(^0.0.0.0 \)//g') && \
    COUNT_BEFORE=$(echo "$hostnames" | sed '/^\s*$/d' | wc -l) && \
    hostnames=$(echo "$hostnames" | sort | uniq | sed '/\(psma01.com.\)\|\(psma02.com.\)\|\(psma03.com.\)\|\(MEZIAMUSSUCEMAQUEUE.SU\)/d') && \
    COUNT_AFTER=$(echo "$hostnames" | sed '/^\s*$/d' | wc -l) && \
    echo "Removed $((COUNT_BEFORE-$COUNT_AFTER)) duplicates from $COUNT_BEFORE hostnames" && \
    COUNT_BEFORE=$(echo "$hostnames" | sed '/^\s*$/d' | wc -l) && \
    hostnames=$(echo "$hostnames" | sed '/\(maxmind.com\)/Id') && \
    COUNT_AFTER=$(echo "$hostnames" | sed '/^\s*$/d' | wc -l) && \
    echo "Removed $((COUNT_BEFORE-$COUNT_AFTER)) entries manually (see Dockerfile)" && \
    for hostname in $hostnames; do echo "local-zone: \""$hostname"\" static" >> blocks-malicious.conf; done && \
    tar -cjf blocks-malicious.conf.bz2 blocks-malicious.conf

FROM alpine:3.8
LABEL maintainer="quentin.mcgaw@gmail.com" \
      description="Runs a DNS server connected to Cloudflare DNS server 1.1.1.1 over TLS" \
      download="5MB" \
      size="12.1MB" \
      ram="13.2MB or 82MB" \
      cpu_usage="Very low to low" \
      github="https://github.com/qdm12/cloudflare-dns-server"
EXPOSE 53/udp
ENV VERBOSITY=1 \
    VERBOSITY_DETAILS=0 \
    BLOCK_MALICIOUS=on
HEALTHCHECK --interval=5m --timeout=15s --start-period=5s --retries=2 CMD if [[ "$(nslookup duckduckgo.com 2>nul)" == "" ]]; then echo "Can't resolve duckduckgo.com"; exit 1; fi
COPY --from=rootanchors /named.root /etc/unbound/root.hints
COPY --from=rootanchors /root.key /etc/unbound/root.key
COPY --from=blocks /blocks-malicious.conf.bz2 /etc/unbound/blocks-malicious.conf.bz2
RUN echo https://alpine.global.ssl.fastly.net/alpine/v3.8/main > /etc/apk/repositories && \
	apk add --update --no-cache -q --progress unbound && \
    rm -rf /var/cache/apk/* /etc/unbound/unbound.conf && \
    echo "# Add Unbound configuration below" > /etc/unbound/include.conf && \
    chown unbound /etc/unbound/root.key
COPY unbound.conf entrypoint.sh /etc/unbound/
RUN chmod +x /etc/unbound/entrypoint.sh
ENTRYPOINT /etc/unbound/entrypoint.sh
