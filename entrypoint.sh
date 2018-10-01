#!/bin/sh

printf "\n ========================================="
printf "\n ========================================="
printf "\n === CLOUDFLARE DNS OVER TLS CONTAINER ==="
printf "\n ========================================="
printf "\n ========================================="
printf "\n == by github.com/qdm12 - Quentin McGaw ==\n\n"

sed -i "s/verbosity: 2/verbosity: $VERBOSITY/g" /etc/unbound/unbound.conf
$(grep -Fq 127.0.0.1 /etc/resolv.conf) || echo "WARNING: The domain name is not set to 127.0.0.1 so the healthcheck will likely be irrelevant!"
unbound -d -v -v
status=$?
printf "\n ========================================="
printf "\n ========================================="
printf "\n Unbound exited with status $status"
printf "\n =========================================\n"