# Cloudflare DNS over TLS Docker container

Docker container running a DNS using Cloudflare **1.1.1.1** DNS over TLS (IPv4 only), with a built-in *healthcheck*

[![Docker Cloudflare DNS](https://github.com/qdm12/cloudflare-dns-server/raw/master/readme/title.png)](https://hub.docker.com/r/qmcgaw/cloudflare-dns-server)

[![Build Status](https://travis-ci.org/qdm12/cloudflare-dns-server.svg?branch=master)](https://travis-ci.org/qdm12/cloudflare-dns-server)
[![Docker Build Status](https://img.shields.io/docker/build/qmcgaw/cloudflare-dns-server.svg)](https://hub.docker.com/r/qmcgaw/cloudflare-dns-server)

[![GitHub last commit](https://img.shields.io/github/last-commit/qdm12/cloudflare-dns-server.svg)](https://github.com/qdm12/cloudflare-dns-server/issues)
[![GitHub commit activity](https://img.shields.io/github/commit-activity/y/qdm12/cloudflare-dns-server.svg)](https://github.com/qdm12/cloudflare-dns-server/issues)
[![GitHub issues](https://img.shields.io/github/issues/qdm12/cloudflare-dns-server.svg)](https://github.com/qdm12/cloudflare-dns-server/issues)

[![Docker Pulls](https://img.shields.io/docker/pulls/qmcgaw/cloudflare-dns-server.svg)](https://hub.docker.com/r/qmcgaw/cloudflare-dns-server)
[![Docker Stars](https://img.shields.io/docker/stars/qmcgaw/cloudflare-dns-server.svg)](https://hub.docker.com/r/qmcgaw/cloudflare-dns-server)
[![Docker Automated](https://img.shields.io/docker/automated/qmcgaw/cloudflare-dns-server.svg)](https://hub.docker.com/r/qmcgaw/cloudflare-dns-server)

[![](https://images.microbadger.com/badges/image/qmcgaw/cloudflare-dns-server.svg)](https://microbadger.com/images/qmcgaw/cloudflare-dns-server)
[![](https://images.microbadger.com/badges/version/qmcgaw/cloudflare-dns-server.svg)](https://microbadger.com/images/qmcgaw/cloudflare-dns-server)

| Download size | Image size | RAM usage | CPU usage |
| --- | --- | --- | --- |
| 4.3MB | 9.58MB | 6MB | Very Low |

It is based on:
- [Alpine 3.7](https://alpinelinux.org)
- [Unbound 1.7.0-r2](https://pkgs.alpinelinux.org/package/edge/main/aarch64/unbound)

Diagrams are shown for router and client-by-client configurations in the [**Connect clients to it**](#connect-clients-to-it) section

## Testing it

```bash
docker run -it --rm -p 53:53/udp qmcgaw/cloudflare-dns-server -vvv
```


Note the `-vvv` to set the verbose level to 3. It defaults to 1 if no command is provided.

See the [Connect clients to it](#connect-clients-to-it) section to finish testing.

## Run it as a daemon

```bash
docker run -d --name=cloudflareTlsDNS -p 53:53/udp qmcgaw/cloudflare-dns-server
```

You can also download  and use [*docker-compose.yml*](https://github.com/qdm12/cloudflare-dns-server/blob/master/docker-compose.yml)

## Connect clients to it

### Option 1: Router (recommended)

*All machines connected to your router will use the 1.1.1.1 encrypted DNS by default*

Configure your router to use the LAN IP address of your Docker host as its primary DNS address.
- Access your router page, usually at [http://192.168.1.1](http://192.168.1.1) and login with your credentials
- Change the DNS settings, which are usually located in *Connection settings / Advanced / DNS server*
- If a secondary fallback DNS address is required, use Cloudflare address **1.1.1.1** without TLS

![](https://github.com/qdm12/cloudflare-dns-server/blob/master/readme/diagram-router.png?raw=true)

### Option 2: Client, one by one

You have to configure each machine connected to your router to use the Docker host as their DNS server.

![](https://github.com/qdm12/cloudflare-dns-server/blob/master/readme/diagram-clients.png?raw=true)

#### Docker containers

Connect other Docker containers by specifying the DNS to be **127.0.0.1**

- Use the argument `--dns=127.0.0.1` with the `docker run` command
- Or modify your *docker-compose.yml* by adding the following to your container description:

    ```yml
    dns:
        - 127.0.0.1
    ```

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

You probably know how to do that. Otherwise you can usually modify the first line of */etc/resolv.conf* by changing the IP address 
of your DNS server.

#### Android

See [this](http://xslab.com/2013/08/how-to-change-dns-settings-on-android/)

#### iOS

See [this](http://www.macinstruct.com/node/558)