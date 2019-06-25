# Cloudflare DNS over TLS Docker container

*DNS caching server connected to DNS over TLS (IPv4) servers with DNSSEC, DNS rebinding protection, built-in Docker healthcheck and malicious IPs + hostnames blocking*

[![Cloudflare DNS over TLS Docker](https://github.com/qdm12/cloudflare-dns-server/raw/master/readme/title.png)](https://hub.docker.com/r/qmcgaw/cloudflare-dns-server)

[![Docker Build Status](https://img.shields.io/docker/build/qmcgaw/cloudflare-dns-server.svg)](https://hub.docker.com/r/qmcgaw/cloudflare-dns-server)

[![GitHub last commit](https://img.shields.io/github/last-commit/qdm12/cloudflare-dns-server.svg)](https://github.com/qdm12/cloudflare-dns-server/commits)
[![GitHub commit activity](https://img.shields.io/github/commit-activity/y/qdm12/cloudflare-dns-server.svg)](https://github.com/qdm12/cloudflare-dns-server/commits)
[![GitHub issues](https://img.shields.io/github/issues/qdm12/cloudflare-dns-server.svg)](https://github.com/qdm12/cloudflare-dns-server/issues)

[![Docker Pulls](https://img.shields.io/docker/pulls/qmcgaw/cloudflare-dns-server.svg)](https://hub.docker.com/r/qmcgaw/cloudflare-dns-server)
[![Docker Stars](https://img.shields.io/docker/stars/qmcgaw/cloudflare-dns-server.svg)](https://hub.docker.com/r/qmcgaw/cloudflare-dns-server)
[![Docker Automated](https://img.shields.io/docker/automated/qmcgaw/cloudflare-dns-server.svg)](https://hub.docker.com/r/qmcgaw/cloudflare-dns-server)

[![Image size](https://images.microbadger.com/badges/image/qmcgaw/cloudflare-dns-server.svg)](https://microbadger.com/images/qmcgaw/cloudflare-dns-server)
[![Image version](https://images.microbadger.com/badges/version/qmcgaw/cloudflare-dns-server.svg)](https://microbadger.com/images/qmcgaw/cloudflare-dns-server)

[![Donate PayPal](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://paypal.me/qdm12)

| Image size | RAM usage | CPU usage |
| --- | --- | --- |
| 26.9MB | 13.2MB to 70MB | Low |

It can be connected to one of all the DNS-over-TLS providers:

- Google
- Quad9
- Quadrant
- CleanBrowsing

<details><summary>Click to show base components</summary><p>

- [Alpine 3.10](https://alpinelinux.org)
- [Unbound 1.9.1-r2](https://pkgs.alpinelinux.org/package/v3.10/main/x86_64/unbound)
- [Files and lists built periodically](https://github.com/qdm12/updated/tree/master/files)
- [bind-tools](https://pkgs.alpinelinux.org/package/v3.10/main/x86_64/bind-tools) for the healthcheck with `nslookup duckduckgo.com 127.0.0.1`

</p></details>

It also uses DNS rebinding protection and DNSSEC Validation:

[![DNSSEC Validation](https://github.com/qdm12/cloudflare-dns-server/blob/master/readme/rootcanary.org.png?raw=true)](https://www.rootcanary.org/test.html)

You can also block additional domains of your choice, amongst other things, see the [Extra section](#Extra)

Diagrams are shown for router and client-by-client configurations in the [**Connect clients to it**](#connect-clients-to-it) section.

## Testing it

1. <details><summary>CLICK IF YOU HAVE AN ARM DEVICE</summary><p>

    - If you have a ARM 32 bit v6 architecture

        ```sh
        docker build -t qmcgaw/cloudflare-dns-server \
        --build-arg BASE_IMAGE=arm32v6/alpine \
        https://github.com/qdm12/cloudflare-dns-server.git
        ```

    - If you have a ARM 32 bit v7 architecture

        ```sh
        docker build -t qmcgaw/cloudflare-dns-server \
        --build-arg BASE_IMAGE=arm32v7/alpine \
        https://github.com/qdm12/cloudflare-dns-server.git
        ```

    - If you have a ARM 64 bit v8 architecture

        ```sh
        docker build -t qmcgaw/cloudflare-dns-server \
        --build-arg BASE_IMAGE=arm64v8/alpine \
        https://github.com/qdm12/cloudflare-dns-server.git
        ```

    </p></details>

1. Run the container

    ```bash
    docker run -it --rm -p 53:53/udp -e VERBOSITY=3 -e VERBOSITY_DETAILS=3 qmcgaw/cloudflare-dns-server
    ```

    More environment variables are described in the [environment variables](#environment-variables) section.

1. Check the log output

    ```bash
    docker logs cloudflare-dns-server
    ```

1. See the [Connect clients to it](#connect-clients-to-it) section to finish testing, and you can refer to the [Verify DNS connection](#verify-dns-connection) section if you want.

## Run it as a daemon

```bash
docker run -d -p 53:53/udp qmcgaw/cloudflare-dns-server
```

or use [docker-compose.yml](https://github.com/qdm12/cloudflare-dns-server/blob/master/docker-compose.yml) with:

```bash
docker-compose up -d
```

More environment variables are described in the [environment variables](#environment-variables) section.

## Environment variables

| Environment variable | Default | Description |
| --- | --- | --- |
| `VERBOSITY` | `1` | From 0 (no log) to 5 (full debug log) |
| `VERBOSITY_DETAILS` | `0` | From 0 to 4 and defaults to 0 (higher means more details) |
| `BLOCK_MALICIOUS` | `on` | `on` or `off`. It blocks malicious IP addresses and malicious hostnames from being resolved. Note that it consumes about 50MB of additional RAM. |
| `BLOCK_NSA` | `off` | `on` or `off`. It blocks NSA hostnames from being resolved. |
| `UNBLOCK` | | comma separated list of hostnames to leave unblocked |
| `LISTENINGPORT` | `53` | UDP port on which the Unbound DNS server should listen to (internally) |
| `PROVIDER` | `cloudflare` | DNS-over-TLS provider. It can be: `google`, `quad9`, `quadrant`, `cleanbrowsing` |

## Connect clients to it

### Option 1: Router (recommended)

Block the UDP 53 outgoing port on your router firewall so that all DNS traffic must go through this container.

*All machines connected to your router will use the 1.1.1.1 encrypted DNS by default*

Configure your router to use the LAN IP address of your Docker host as its primary DNS address.

- Access your router page, usually at [http://192.168.1.1](http://192.168.1.1) and login with your credentials
- Change the DNS settings, which are usually located in *Connection settings / Advanced / DNS server*
- If a secondary fallback DNS address is required, use a dull ip address such as the router's IP 192.168.1.1 to force traffic to only go through this container

![](https://github.com/qdm12/cloudflare-dns-server/blob/master/readme/diagram-router.png?raw=true)

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
    image: alpine:3.10
    network_mode: bridge
    dns:
      - 127.0.0.1
```

If the containers are in the same virtual network, you can simply set the `dns` to the LAN IP address of the DNS container (i.e. `10.0.0.5`)

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

## Extra

### Block domains of your choice

1. Create a file on your host `include.conf`
1. Write the following to the file to block *youtube.com* for example:

    ```txt
    local-zone: "youtube.com" static
    ```

1. Change the ownership and permissions of `include.conf`:

    ```bash
    chown 1000:1000 include.conf
    chmod 400 include.conf
    ```

1. Launch the Docker container with:

    ```bash
    docker run -it --rm -p 53:53/udp -v $(pwd)/include.conf:/etc/unbound/include.conf  qmcgaw/cloudflare-dns-server
    ```

### Build the image yourself

```bash
docker build -t qmcgaw/cloudflare-dns-server https://github.com/qdm12/cloudflare-dns-server.git
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

- [ ] Build Unbound binary at image build stage
