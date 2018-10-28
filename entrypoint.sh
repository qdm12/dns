#!/bin/sh

printf "\n ========================================="
printf "\n ========================================="
printf "\n === CLOUDFLARE DNS OVER TLS CONTAINER ==="
printf "\n ========================================="
printf "\n ========================================="
printf "\n == by github.com/qdm12 - Quentin McGaw ==\n\n"

printf "\nUnbound version: $(unbound -h | grep "Version" | cut -d" " -f2)"
printf "\nVerbosity level set to $VERBOSITY"
printf "\nVerbosity details level set to $VERBOSITY_DETAILS"
printf "\nMalicious hostnames and ips blocking is $BLOCK_MALICIOUS\n"
[[ "$VERBOSITY" == "" ]] || sed -i "s/verbosity: 0/verbosity: $VERBOSITY/g" /etc/unbound/unbound.conf
$(grep -Fq 127.0.0.1 /etc/resolv.conf) || echo "WARNING: The domain name is not set to 127.0.0.1 so the healthcheck will likely be irrelevant!"
[[ "$VERBOSITY_DETAILS" == "" ]] || [[ "$VERBOSITY_DETAILS" == "0" ]] || ARGS=-$(for i in `seq $VERBOSITY_DETAILS`; do printf "v"; done)
if [ "$BLOCK_MALICIOUS" = "on" ] && [ ! -f /etc/unbound/blocks-malicious.conf ]; then
    printf "Extracting malicious hostnames archive..."
    tar -xjf /etc/unbound/malicious-hostnames.bz2 -C /etc/unbound/
    printf "DONE\n"
    printf "Extracting malicious IPs archive..."
    tar -xjf /etc/unbound/malicious-ips.bz2 -C /etc/unbound/
    printf "DONE\n"
    printf "Building blocks-malicious.conf for Unbound..."
    while read hostname; do
        echo "local-zone: \""$hostname"\" static" >> /etc/unbound/blocks-malicious.conf
    done < /etc/unbound/malicious-hostnames
    while read ip; do
        echo "private-address: $ip" >> /etc/unbound/blocks-malicious.conf
    done < /etc/unbound/malicious-ips
    printf "$(cat /etc/unbound/malicious-hostnames | wc -l ) malicious hostnames and $(cat /etc/unbound/malicious-ips | wc -l) malicious IP addresses added\n"
    rm -f /etc/unbound/malicious-hostnames* /etc/unbound/malicious-ips*
else
    touch /etc/unbound/blocks-malicious.conf
fi
unbound -d $ARGS
status=$?
printf "\n ========================================="
printf "\n Unbound exited with status $status"
printf "\n =========================================\n"