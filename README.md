# DNS over TLS upstream server Docker container

*DNS over TLS upstream server connected to DNS over TLS (IPv4) servers with DNSSEC, DNS rebinding protection, built-in Docker healthcheck and fine grain IPs + hostnames blocking*

**ANNOUCEMENT**: *Total rewrite in Go: see the new features [below](#Features)* (in case something break, use the image with tag `:shell`)

[![Cloudflare DNS over TLS Docker](https://github.com/qdm12/cloudflare-dns-server/raw/master/readme/title.png)](https://hub.docker.com/r/qmcgaw/cloudflare-dns-server)

[![Build Status](https://travis-ci.org/qdm12/cloudflare-dns-server.svg?branch=master)](https://travis-ci.org/qdm12/cloudflare-dns-server)
[![Docker Pulls](https://img.shields.io/docker/pulls/qmcgaw/cloudflare-dns-server.svg)](https://hub.docker.com/r/qmcgaw/cloudflare-dns-server)
[![Docker Stars](https://img.shields.io/docker/stars/qmcgaw/cloudflare-dns-server.svg)](https://hub.docker.com/r/qmcgaw/cloudflare-dns-server)

[![GitHub last commit](https://img.shields.io/github/last-commit/qdm12/cloudflare-dns-server.svg)](https://github.com/qdm12/cloudflare-dns-server/commits)
[![GitHub commit activity](https://img.shields.io/github/commit-activity/y/qdm12/cloudflare-dns-server.svg)](https://github.com/qdm12/cloudflare-dns-server/commits)
[![GitHub issues](https://img.shields.io/github/issues/qdm12/cloudflare-dns-server.svg)](https://github.com/qdm12/cloudflare-dns-server/issues)

[![Image size](https://images.microbadger.com/badges/image/qmcgaw/cloudflare-dns-server.svg)](https://microbadger.com/images/qmcgaw/cloudflare-dns-server)
[![Image version](https://images.microbadger.com/badges/version/qmcgaw/cloudflare-dns-server.svg)](https://microbadger.com/images/qmcgaw/cloudflare-dns-server)
[![Join Slack channel](https://img.shields.io/badge/slack-@qdm12-yellow.svg?logo=slack)](https://join.slack.com/t/qdm12/shared_invite/enQtODMwMDQyMTAxMjY1LTU1YjE1MTVhNTBmNTViNzJiZmQwZWRmMDhhZjEyNjVhZGM4YmIxOTMxOTYzN2U0N2U2YjQ2MDk3YmYxN2NiNTc)

## Features

It can be connected to one or more of the following DNS-over-TLS providers:

- [Cloudflare](https://developers.cloudflare.com/1.1.1.1/dns-over-tls/)
- [Google](https://developers.google.com/speed/public-dns/docs/dns-over-tls)
- [Quad9](https://www.quad9.net/faq/#Does_Quad9_support_DNS_over_TLS)
- [SecureDNS](https://securedns.eu)
- [LibreDNS](https://libredns.gr)
- [Quadrant](https://quadrantsec.com/about/blog/quadrants_public_dns_resolver_with_tls_https_support/)
- [CleanBrowsing](https://cleanbrowsing.org/guides/dnsovertls)

<details><summary>New features added</summary><p>

- Smaller image of 41.3MB and faster boot up
- Download cryptographic files and needed block files at container start time
- Block hostnames and IP addresses for 3 categories: malicious, surveillance and ads
- Block custom hostnames and IP addresses using environment variables
- Added environment variable `VALIDATION_LOGLEVEL` for Unbound

</p></details>

<details><summary>Click to show base components</summary><p>

- [Alpine 3.11](https://alpinelinux.org)
- [Unbound 1.9.6](https://nlnetlabs.nl/downloads/unbound) ~built from source~ (from Alpine packages)
- [Files and lists built periodically](https://github.com/qdm12/updated/tree/master/files)
- Go static binary built from this source

</p></details>

- Compatible with amd64, i686 (32 bit), **ARM** 64 bit, ARM 32 bit v6 and v7, ppc64le and even that s390x 🎆
- DNS rebinding protection
- DNSSEC Validation

[![DNSSEC Validation](https://github.com/qdm12/cloudflare-dns-server/blob/master/readme/rootcanary.org.png?raw=true)](https://www.rootcanary.org/test.html)

- Split-horizon DNS (randomly pick one of the DoT providers)
- Optional hostnames resolution and IPs blocking
    - Malicious
    - Surveillance
    - Ads
    - Custom

Diagrams are shown for router and client-by-client configurations in the [**Connect clients to it**](#connect-clients-to-it) section.

## Setup

1. Launch the container with

    ```sh
    docker run -d -p 53:53/udp qmcgaw/cloudflare-dns-server
    ```

    You can also use [docker-compose.yml](https://github.com/qdm12/cloudflare-dns-server/blob/master/docker-compose.yml) with:

    ```sh
    docker-compose up -d
    ```

    More environment variables are described in the [environment variables](#environment-variables) section.

1. See the [Connect clients to it](#connect-clients-to-it) section, you can also refer to the [Verify DNS connection](#verify-dns-connection) section if you want.

## Environment variables

| Environment variable | Default | Description |
| --- | --- | --- |
| `PROVIDERS` | `cloudflare` | Comma separated list of DNS-over-TLS providers from `cloudflare`, `google`, `quad9`, `quadrant`, `cleanbrowsing`, `securedns` and `libredns` |
| `VERBOSITY` | `1` | From 0 (no log) to 5 (full debug log) |
| `VERBOSITY_DETAILS` | `1` | From 0 to 4 (higher means more details) |
| `BLOCK_MALICIOUS` | `on` | `on` or `off`, to block malicious IP addresses and malicious hostnames from being resolved |
| `BLOCK_SURVEILLANCE` | `on` | `on` or `off`, to block surveillance IP addresses and hostnames from being resolved |
| `BLOCK_ADS` | `off` | `on` or `off`, to block ads IP addresses and hostnames from being resolved |
| `BLOCK_HOSTNAMES` |  | comma separated list of hostnames to block from being resolved |
| `BLOCK_IPS` |  | comma separated list of IPs to block from being returned to clients |
| `UNBLOCK` | | comma separated list of hostnames to leave unblocked |
| `LISTENINGPORT` | `53` | UDP port on which the Unbound DNS server should listen to (internally) |
| `CACHING` | `on` | `on` or `off`. It can be useful if you have another DNS (i.e. Pihole) doing the caching as well on top of this container |
| `PRIVATE_ADDRESS` | All IPv4 and IPv6 CIDRs private ranges | Comma separated list of CIDRs or single IP addresses. Note that the default setting prevents DNS rebinding |

## Connect clients to it

### Option 1: Router (recommended)

*All machines connected to your router will use the 1.1.1.1 encrypted DNS by default*

Configure your router to use the LAN IP address of your Docker host as its primary DNS address.

- Access your router page, usually at [http://192.168.1.1](http://192.168.1.1) and login with your credentials
- Change the DNS settings, which are usually located in *Connection settings / Advanced / DNS server*
- If a secondary fallback DNS address is required, use a dull ip address such as the router's IP 192.168.1.1 to force traffic to only go through this container

![](https://github.com/qdm12/cloudflare-dns-server/blob/master/readme/diagram-router.png?raw=true)

To ensure network clients cannot use another DNS, you might want to

- Block the outbound UDP 53 port on your router firewall
- Block the outbound TCP 853 port on your router firewall, **except from your Docker host**
- If you have *Deep packet inspection* on your router, block DNS over HTTPs on port TCP 443

### Option 2: Client, one by one

You have to configure each machine connected to your router to use the Docker host as their DNS server.

![](https://github.com/qdm12/cloudflare-dns-server/blob/master/readme/diagram-clients.png?raw=true)

#### Docker containers

Connect other Docker containers by specifying the DNS to be the host IP address `127.0.0.1`:

```bash
docker run -it --rm --dns=127.0.0.1 alpine
```

For *docker-compose.yml*:

```yml
version: '3'
services:
  test:
    image: alpine:3.11
    network_mode: bridge
    dns:
      - 127.0.0.1
```

If the containers are in the same Docker network, you can simply set the `dns` to the LAN IP address of the DNS container (i.e. `10.0.0.5`)

#### Windows

1. Open the control panel and follow the instructions shown on the screenshots below.

![](https://github.com/qdm12/cloudflare-dns-server/blob/master/readme/windows1.png?raw=true)

![](https://github.com/qdm12/cloudflare-dns-server/blob/master/readme/windows2.png?raw=true)

![](https://github.com/qdm12/cloudflare-dns-server/blob/master/readme/windows3.png?raw=true)

![](https://github.com/qdm12/cloudflare-dns-server/blob/master/readme/windows4.png?raw=true)

![](https://github.com/qdm12/cloudflare-dns-server/blob/master/readme/windows5.png?raw=true)

Enter the IP Address of your Docker host as the **Preferred DNS server** (`192.168.1.210` in my case)
You can set the Cloudflare DNS server address 1.1.1.1 as an alternate DNS server although you might want to
leave this blank so that no domain name request is in plaintext.

![](https://github.com/qdm12/cloudflare-dns-server/blob/master/readme/windows6.png?raw=true)

![](https://github.com/qdm12/cloudflare-dns-server/blob/master/readme/windows7.png?raw=true)

When closing, Windows should try to identify any potential problems.
If everything is fine, you should see the following message:

![](https://github.com/qdm12/cloudflare-dns-server/blob/master/readme/windows8.png?raw=true)

#### Mac OS

Follow the instructions at [https://support.apple.com/kb/PH25577](https://support.apple.com/kb/PH25577)

#### Linux

You probably know how to do that. Otherwise you can usually modify the first line of */etc/resolv.conf* by changing the IP address of your DNS server.

#### Android

See [this](http://xslab.com/2013/08/how-to-change-dns-settings-on-android/)

#### iOS

See [this](http://www.macinstruct.com/node/558)

### Build the image yourself

- Build the latest Docker image
    - With `git`

        ```sh
        docker build -t qmcgaw/cloudflare-dns-server https://github.com/qdm12/cloudflare-dns-server.git
        ```

    - With `wget` and `unzip`

        ```sh
        wget -q "https://github.com/qdm12/cloudflare-dns-server/archive/master.zip"
        unzip -q "master.zip"
        cd *-master
        docker build -t qmcgaw/cloudflare-dns-server .
        cd .. && rm -r master.zip *-master
        ```

- Build an older Docker image (you need `wget` and `unzip`)
    1. Go to [the commits](https://github.com/qdm12/cloudflare-dns-server/commits/master) and find which commit you want to build for
    1. You can click on the clipboard next to the commit, in example you pick the commit `da6dbb2ff21c0af4cee93fdb92415aee167f7fd7`
    1. Open a terminal and set `COMMIT=da6dbb2ff21c0af4cee93fdb92415aee167f7fd7`
    1. Download the code for this commit and build the Docker image, either:
        - With `git`

            ```sh
            git clone https://github.com/qdm12/cloudflare-dns-server.git temp
            cd temp
            git reset --hard $COMMIT
            docker build -t qmcgaw/cloudflare-dns-server .
            cd .. && rm -r temp
            ```

        - With `wget` and `unzip`

            ```sh
            wget -q "https://github.com/qdm12/cloudflare-dns-server/archive/$COMMIT.zip"
            unzip -q "$COMMIT.zip"
            cd *-$COMMIT
            docker build -t qmcgaw/cloudflare-dns-server .
            cd .. && rm -r "$COMMIT.zip" *-$COMMIT
            ```

### Firewall considerations

This container requires the following connections:

- UDP 53 Inbound (only if used externally)
- TCP 853 Outbound to 1.1.1.1 and 1.0.0.1

### Verify DNS connection

1. Verify that you use Cloudflare DNS servers: [https://www.dnsleaktest.com](https://www.dnsleaktest.com) with the Standard or Extended test
1. Verify that DNS SEC is enabled: [https://en.internet.nl/connection](https://en.internet.nl/connection)

Note that [https://1.1.1.1/help](https://1.1.1.1/help) does not work as the container is not a client to Cloudflare servers but a forwarder intermediary. Hence https://1.1.1.1/help does not detect a direct connection to them.

## TO DOs

- [ ] Periodic SHUP signal to reload block lists
- [x] Build Unbound binary at image build stage
    - [ ] smaller static binary
    - [ ] Bundled with Go static binary on a Scratch image
- [ ] Branch with Pihole bundled
