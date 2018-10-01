#!/bin/sh

printf "\n ========================================="
printf "\n ========================================="
printf "\n === CLOUDFLARE DNS OVER TLS CONTAINER ==="
printf "\n ========================================="
printf "\n ========================================="
printf "\n == by github.com/qdm12 - Quentin McGaw ==\n\n"

[[ "$VERBOSITY" == "" ]] || sed -i "s/verbosity: 0/verbosity: $VERBOSITY/g" /etc/unbound/unbound.conf
$(grep -Fq 127.0.0.1 /etc/resolv.conf) || echo "WARNING: The domain name is not set to 127.0.0.1 so the healthcheck will likely be irrelevant!"
[[ "$VERBOSITY_DETAILS" == "" ]] || [[ "$VERBOSITY_DETAILS" == "0" ]] || ARGS=-$(for i in `seq $VERBOSITY_DETAILS`; do printf "v"; done)
unbound -d $ARGS
status=$?
printf "\n ========================================="
printf "\n Unbound exited with status $status"
printf "\n =========================================\n"