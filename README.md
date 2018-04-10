# Cloudflare DNS Docker container

Docker container running a DNS using Cloudflare **1.1.1.1** DNS over TLS

[![Docker Cloudflare DNS](https://github.com/qdm12/cloudflare-dns-server/raw/master/readme/title.png)](https://hub.docker.com/r/qmcgaw/cloudflare-dns-server)

[![Build Status](https://travis-ci.org/qdm12/cloudflare-dns-server.svg?branch=master)](https://travis-ci.org/qdm12/cloudflare-dns-server)

[![](https://images.microbadger.com/badges/image/qmcgaw/cloudflare-dns-server.svg)](https://microbadger.com/images/qmcgaw/cloudflare-dns-server)
[![](https://images.microbadger.com/badges/version/qmcgaw/cloudflare-dns-server.svg)](https://microbadger.com/images/qmcgaw/cloudflare-dns-server)

| Download size | Image size | RAM usage | CPU usage |
| --- | --- | --- | --- |
| 4.3MB | 9.58MB | 6MB | Very Low |

It is based on:
- Alpine 3.7
- Unbound 1.6.7

## Running it

1. Run it with:

    ```bash
    docker run -d --name=cloudflareTlsDNS -p 53:53/udp qmcgaw/cloudflare-dns-server
    ```

1. Configure your router to use the LAN IP address of your Docker host as its primary DNS address.
If a secondary DNS address is required, use cloudfare address directly as a fallback 1.1.1.1