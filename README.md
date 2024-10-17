# DNS over TLS or HTTPs forwarding resolver

Resolver communicating with public DNS recursive servers over encrypted channels with TLS or HTTPs.
It also does **caching**, **filtering**, **split-horizon DNS**, **IPv6**, **Prometheus metrucs**.
It's fully coded in Go and is a single and cross platform binary program.

**Announcement**: *I am currently working on a DNSSEC validator implementation to reach feature parity with the v1.x.x image using Unbound*

**The `:v2.0.0-beta` Docker image breaks compatibility with previous images based on v1.x.x versions**

[![Title](https://github.com/qdm12/dns/raw/master/readme/title.png)](https://hub.docker.com/r/qmcgaw/dns)

[![Build status](https://github.com/qdm12/dns/actions/workflows/build.yml/badge.svg)](https://github.com/qdm12/dns/actions/workflows/build.yml)

[![dockeri.co](https://dockeri.co/image/qmcgaw/dns)](https://hub.docker.com/r/qmcgaw/dns)
[![dockeri.co](https://dockeri.co/image/qmcgaw/cloudflare-dns-server)](https://hub.docker.com/r/qmcgaw/cloudflare-dns-server)

![Last release](https://img.shields.io/github/release/qdm12/dns?label=Last%20release)
![Last Docker tag](https://img.shields.io/docker/v/qmcgaw/dns?sort=semver&label=Last%20Docker%20tag)
[![Last release size](https://img.shields.io/docker/image-size/qmcgaw/dns?sort=semver&label=Last%20released%20image)](https://hub.docker.com/r/qmcgaw/dns/tags?page=1&ordering=last_updated)
![GitHub last release date](https://img.shields.io/github/release-date/qdm12/dns?label=Last%20release%20date)
![Commits since release](https://img.shields.io/github/commits-since/qdm12/dns/latest?sort=semver)

[![Latest size](https://img.shields.io/docker/image-size/qmcgaw/dns/latest?label=Latest%20image)](https://hub.docker.com/r/qmcgaw/dns/tags)

[![GitHub last commit](https://img.shields.io/github/last-commit/qdm12/dns.svg)](https://github.com/qdm12/dns/commits/main)
[![GitHub commit activity](https://img.shields.io/github/commit-activity/y/qdm12/dns.svg)](https://github.com/qdm12/dns/graphs/contributors)
[![GitHub closed PRs](https://img.shields.io/github/issues-pr-closed/qdm12/dns.svg)](https://github.com/qdm12/dns/pulls?q=is%3Apr+is%3Aclosed)
[![GitHub issues](https://img.shields.io/github/issues/qdm12/dns.svg)](https://github.com/qdm12/dns/issues)
[![GitHub closed issues](https://img.shields.io/github/issues-closed/qdm12/dns.svg)](https://github.com/qdm12/dns/issues?q=is%3Aissue+is%3Aclosed)

[![Lines of code](https://img.shields.io/tokei/lines/github/qdm12/dns)](https://github.com/qdm12/dns)
![Code size](https://img.shields.io/github/languages/code-size/qdm12/dns)
![GitHub repo size](https://img.shields.io/github/repo-size/qdm12/dns)
![Go version](https://img.shields.io/github/go-mod/go-version/qdm12/dns)

[![MIT](https://img.shields.io/github/license/qdm12/dns)](https://github.com/qdm12/dns/master/LICENSE)
![Visitors count](https://visitor-badge.laobi.icu/badge?page_id=dns.readme)

## Features

- It can be connected to one or more of the following public resolvers over both DNS over TLS and DNS over HTTPS:
  - [Cloudflare](https://developers.cloudflare.com/1.1.1.1/dns-over-tls/)
  - [Google](https://developers.google.com/speed/public-dns/docs/dns-over-tls)
  - [LibreDNS](https://libredns.gr)
  - [OpenDNS](https://support.opendns.com/hc/en-us/articles/360038086532-Using-DNS-over-HTTPS-DoH-with-OpenDNS)
  - [Quad9](https://www.quad9.net/faq/#Does_Quad9_support_DNS_over_TLS)
  - [Quadrant](https://quadrantsec.com/about/blog/quadrants_public_dns_resolver_with_tls_https_support/)
  - [CleanBrowsing](https://cleanbrowsing.org/guides/dnsovertls)
  - [CIRA Canadian Shield](https://www.cira.ca/cybersecurity-services/canadian-shield)
- Random split-horizon DNS (an upstream resolver is picked at random for every request received)
- Hostnames and IP addresses filtering 🛑
  - for 3 categories: malicious, surveillance and ads
  - auto-update [block lists](https://github.com/qdm12/files) periodically with minimal downtime
  - Specify custom hostnames and IP addresses
- DNS rebinding protection
- [Prometheus Metrics](https://github.com/qdm12/dns/blob/v2.0.0-beta/readme/metrics)
- Container specific features 🐋
  - Tiny **16MB** Docker image (uncompressed, amd64) based on the empty image [scratch](https://hub.docker.com/_/scratch)
  - Cross CPU architecture support: amd64, i686 (32 bit), **ARM** 64 bit, ARM 32 bit v7 and v6, ppc64le, s390x and riscv64
  - Running without root

Diagrams are shown for router and client-by-client configurations in the [**Connect clients to it**](#connect-clients-to-it) section.

## Setup

### Container setup

```sh
docker run -d -p 53:53/udp -p 53:53/tcp qmcgaw/dns:v2.0.0-beta
```

You can also use [docker-compose.yml](https://github.com/qdm12/dns/blob/v2.0.0-beta/docker-compose.yml) with:

```sh
docker-compose up -d
```

The image is also available as `ghcr.io/qdm12/dns:v2.0.0-beta`.

If you run an old Docker version or Kernel, you might want to run the container as root with `--user="0"` (see [this issue](https://github.com/qdm12/dns/issues/79) for context).

### Binary program setup

🚧 **waiting for release v2.0.0**

Download the prebuilt binary for your platform from the **Assets** section of the last release on the [releases page](https://github.com/qdm12/dns/releases).

If you run on Linux or OSX, make sure to make it executable with `chmod +x dns`.

You can then run it by clicking on it or in your terminal with `./dns`.

### Kubernetes setup

See the [KUBERNETES.md document](https://github.com/qdm12/dns/blob/v2.0.0-beta/KUBERNETES.md).

### Further setup

- [Settings](#settings) ⚙️ - environment variables and/or flags.
- [Connect clients to it](#connect-clients-to-it)
- [Metrics setup](https://github.com/qdm12/dns/blob/v2.0.0-beta/readme/metrics) 〽️
- [Verify DNS connection](#verify-dns-connection)

## Settings

The following table lists all environment variables available.
**For each variable exists a corresponding CLI flag** with the same name but in lowercase and with underscores replaced by dashes.
For example, the environment variable `UPSTREAM_TYPE` corresponds to the CLI flag `--upstream-type`.

| Environment variable | Default | Description |
| --- | --- | --- |
| `UPSTREAM_TYPE` | `DoT` | Upstream DNS connection type: `DoT` for DNS over TLS or `DoH` for DNS over HTTPS |
| `DOT_RESOLVERS` | `cloudflare,google` | Comma separated list of DNS-over-TLS resolver providers from `cira family`, `cira private`, `cira protected`, `cleanbrowsing adult`, `cleanbrowsing family`, `cleanbrowsing security`, `cloudflare`, `cloudflare family`, `cloudflare security`, `google`, `libredns`, `quad9`, `quad9 secured`, `quad9 unsecured` and `quadrant` |
| `DOH_RESOLVERS` | `cloudflare,google` | Comma separated list of DNS-over-HTTPS resolver providers from `cira family`, `cira private`, `cira protected`, `cleanbrowsing adult`, `cleanbrowsing family`, `cleanbrowsing security`, `cloudflare`, `cloudflare family`, `cloudflare security`, `google`, `libredns`, `quad9`, `quad9 secured`, `quad9 unsecured` and `quadrant` |
| `DOT_TIMEOUT` | `3s` | DNS over TLS dial timeout |
| `DOH_TIMEOUT` | `3s` | DNS over HTTPs exchange timeout |
| `BLOCK_MALICIOUS` | `on` | `on` or `off`, to block malicious IP addresses and malicious hostnames from being resolved |
| `BLOCK_SURVEILLANCE` | `off` | `on` or `off`, to block surveillance IP addresses and hostnames from being resolved |
| `BLOCK_ADS` | `off` | `on` or `off`, to block ads IP addresses and hostnames from being resolved |
| `BLOCK_HOSTNAMES` |  | comma separated list of hostnames to block from being resolved |
| `ALLOWED_HOSTNAMES` | | comma separated list of hostnames to leave unblocked |
| `ALLOWED_IPS` | | comma separated list of IP addresses to leave unblocked |
| `ALLOWED_CIDRS` | | comma separated list of IP networks (CIDRs) to leave unblocked |
| `BLOCK_IPS` |  | comma separated list of IPs to block from being returned to clients |
| `BLOCK_CIDRS` |  | comma separated list of IP networks (CIDRs) to block from being returned to clients |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warning` or `error` |
| `LOG_CALLER` | `hidden` | `hidden` or `short` |
| `MIDDLEWARE_LOG_ENABLED` | `off` | `on` or `off` |
| `MIDDLEWARE_LOG_DIRECTORY` | `/var/log/dns/` | Any valid file path |
| `MIDDLEWARE_LOG_REQUESTS` | `off` | `on` or `off` to log DNS requests to the file path specified |
| `MIDDLEWARE_LOG_RESPONSES` | `off` | `on` or `off` to log DNS responses to the file path specified |
| `DOT_TIMEOUT` | `3s` | Duration string to specify the query timeout for DNS over TLS |
| `DOH_TIMEOUT` | `3s` | Duration string to specify the query timeout for DNS over HTTPS |
| `LISTENING_ADDRESS` | `:53` | DNS server listening address |
| `CACHE_TYPE` | `lru` | `lru` or `noop`. LRU caches DNS responses by least recently used |
| `CACHE_LRU_MAX_ENTRIES` | `10000` | Number of elements to keep in the LRU cache. |
| `METRICS_TYPE` | `noop` | `noop` or `prometheus` |
| `METRICS_PROMETHEUS_ADDRESS` | `:9090` | HTTP Prometheus server listening address |
| `METRICS_PROMETHEUS_SUBSYSTEM` | `dns` | Prometheus metrics prefix/subsystem |
| `MIDDLEWARE_LOCALDNS_ENABLED` | `on` | Enable or disable the local DNS middleware |
| `MIDDLEWARE_LOCALDNS_RESOLVERS` | Local DNS servers | Comma separated list of local DNS resolvers to use for local names DNS requests |
| `MIDDLEWARE_SUBSTITUTER_SUBSTITUTIONS` | | JSON encoded list of substitutions. For example `[{"name":"github.com","ips":["1.2.3.4"]}]`. You can also specify the `type`, `class` and `ttl`, where they default respectively to `A`/`AAAA`, `IN` and `300`. |
| `CHECK_DNS` | `on` | `on` or `off`. Check resolving github.com using `127.0.0.1:53` at start |
| `UPDATE_PERIOD` | `24h` | Period to update block lists and restart Unbound. Set to `0` to disable. |

## Migrate

The `v2.x.x` version of the image (starting with `v2.0.0-beta`) is a complete rewrite from scratch in Go.

There are several non-compatible changes between the `v1` and `v2` images:

- The following environment variables are now unused: `PRIVATE_ADDRESS`, `IPV4`, `IPV6`. The program logs an explanation if any of these is set when running a v2.x.x image.
- The following environment variables are now replaced: `LISTENINGPORT`, `PROVIDERS`, `PROVIDER`, `CACHING`, `UNBLOCK`, `CHECK_UNBOUND`, `VERBOSITY`, `VERBOSITY_DETAILS`, `VALIDATION_LOGLEVEL`. The program logs an explanation if any of these is set when running a v2.x.x image.
- You can no longer bind mount an Unbound configuration file
- Caching is enabled by default (in memory LRU cache with up to 10,000 items)

## Golang API

If you want to use the Go code, you can see tiny [examples](examples) of DoT and DoH resolvers and servers using the API developed. You can also implement your interfaces to pass as settings to existing constructors to further customize the behavior of the program.

The Go API exposed to the public (`pkg/` directory) will stay **stable and compatible** for a long time and there is no reason so far to change it.

## Connect clients to it

### Option 1: Router (recommended)

All machines connected to your router will use the DNS server container by default.

Configure your router to use the LAN IP address of your Docker host as its primary DNS address.

- Access your router page, usually at [http://192.168.1.1](http://192.168.1.1) and login with your credentials
- Change the DNS settings, which are usually located in *Connection settings / Advanced / DNS server*
- If a secondary fallback DNS address is required, use a dull ip address such as the router's IP 192.168.1.1 to force traffic to only go through this container

![Diagram router](https://github.com/qdm12/dns/blob/master/readme/diagram-router.png?raw=true)

To ensure network clients cannot use another DNS, you might want to

- Block the outbound UDP 53 port on your router firewall
- Block the outbound TCP 53 port on your router firewall
- Block the outbound TCP 853 port on your router firewall, **except from your Docker host**
- If you have *Deep packet inspection* on your router, block DNS over HTTPs on port TCP 443

### Option 2: Client, one by one

You have to configure each machine connected to your router to use the Docker host as their DNS server.

![Diagram clients](https://github.com/qdm12/dns/blob/master/readme/diagram-clients.png?raw=true)

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
    image: alpine
    network_mode: bridge
    dns:
      - 127.0.0.1
```

If the containers are in the same Docker network, you can simply set the `dns` to the LAN IP address of the DNS container (i.e. `10.0.0.5`)

#### Windows

1. Open the control panel and follow the instructions shown on the screenshots below.

![Windows screenshot 1](https://github.com/qdm12/dns/blob/master/readme/windows1.png?raw=true)

![Windows screenshot 2](https://github.com/qdm12/dns/blob/master/readme/windows2.png?raw=true)

![Windows screenshot 3](https://github.com/qdm12/dns/blob/master/readme/windows3.png?raw=true)

![Windows screenshot 4](https://github.com/qdm12/dns/blob/master/readme/windows4.png?raw=true)

![Windows screenshot 5](https://github.com/qdm12/dns/blob/master/readme/windows5.png?raw=true)

Enter the IP Address of your Docker host as the **Preferred DNS server** (`192.168.1.210` in my case)
You can set the Cloudflare DNS server address 1.1.1.1 as an alternate DNS server although you might want to
leave this blank so that no domain name request is in plaintext.

![Windows screenshot 6](https://github.com/qdm12/dns/blob/master/readme/windows6.png?raw=true)

![Windows screenshot 7](https://github.com/qdm12/dns/blob/master/readme/windows7.png?raw=true)

When closing, Windows should try to identify any potential problems.
If everything is fine, you should see the following message:

![Windows screenshot 8](https://github.com/qdm12/dns/blob/master/readme/windows8.png?raw=true)

#### Mac OS

Follow the instructions at [https://support.apple.com/kb/PH25577](https://support.apple.com/kb/PH25577)

#### Linux

You probably know how to do that. Otherwise you can usually modify the first line of */etc/resolv.conf* by changing the IP address of your DNS server.

#### Android

See [this](http://xslab.com/2013/08/how-to-change-dns-settings-on-android/)

#### iOS

See [this](http://www.macinstruct.com/node/558)

### Verify DNS connection

1. Verify that you use Cloudflare DNS servers: [https://www.dnsleaktest.com](https://www.dnsleaktest.com) with the Standard or Extended test
1. Verify that DNS SEC is enabled: [https://en.internet.nl/connection](https://en.internet.nl/connection)

Note that [https://1.1.1.1/help](https://1.1.1.1/help) does not work as the container is not a client to Cloudflare servers but a forwarder intermediary. Hence [https://1.1.1.1/help](https://1.1.1.1/help) does not detect a direct connection to them.

## Development

### Development setup

#### Using VSCode and Docker

1. Install [Docker](https://docs.docker.com/install/)
    - On Windows, share a drive with Docker Desktop and have the project on that partition
    - On OSX, share your project directory with Docker Desktop
1. With [Visual Studio Code](https://code.visualstudio.com/download), install the [dev containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
1. In Visual Studio Code, press on `F1` and select `Dev Containers: Open Folder in Container...`
1. Your dev environment is ready to go!... and it's running in a container :+1:

#### Locally

1. Install [Go](https://golang.org/dl/), [Docker](https://www.docker.com/products/docker-desktop) and [Git](https://git-scm.com/downloads)
1. Install dependencies

    ```sh
    go mod download
    ```

1. Install [golangci-lint](https://github.com/golangci/golangci-lint#install)

### Commands available

```sh
# Build the binary
go build ./cmd/dns/main.go
# Test the code
go test ./...
# Lint the code
golangci-lint run
# Build the Docker image
docker build -t qmcgaw/dns .
```

See [Contributing](.github/CONTRIBUTING.md) for more information on how to contribute to this repository.
