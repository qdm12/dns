#!/bin/sh

printf " =========================================\n"
printf " =========================================\n"
printf " === CLOUDFLARE DNS OVER TLS CONTAINER ===\n"
printf " =========================================\n"
printf " =========================================\n"
printf " == by github.com/qdm12 - Quentin McGaw ==\n\n"

user=$(whoami)
printf "Running as $user\n"
if [ "$user" = "nonrootuser" ]; then 
  sed -i 's/username: "nonrootuser"/username: ""/' /etc/unbound/unbound.conf
fi
test -r /etc/unbound/include.conf
if [ $? != 0 ] && [ "$user" = "nonrootuser" ]; then
  printf "Please mount the file by setting its ownership and permissions with:\n"
  printf "  chown 1000:1000 include.conf && chmod 400 include.conf\n"
  exit 1
fi
if [ "$(echo $VERBOSITY | grep -E '^[0-9]+$')" = "" ]; then VERBOSITY=1; fi
if [ "$(echo $VERBOSITY_DETAILS | grep -E '^[0-9]+$')" = "" ]; then VERBOSITY_DETAILS=0; fi
if [ "$(echo $BLOCK_MALICIOUS | grep -E '^[0-9]+$')" = "" ]; then BLOCK_MALICIOUS=on; fi
if [ "$(echo $LISTENINGPORT | grep -E '^[0-9]+$')" = "" ]; then LISTENINGPORT=1053; fi
if [ $LISTENINGPORT -lt 1 ] || [ $LISTENINGPORT -gt 65535 ]; then
  printf "Listening port $LISTENINGPORT must be between port 1 and port 65535\n"
  exit 1
fi
if [ $LISTENINGPORT -lt 1024 ] && [ "$user" = "nonrootuser" ]; then
  printf "Listening port $LISTENINGPORT must be higher than well known ports (1-1023) for a non root user\n"
  printf "You can either change the LISTENINGPORT or run the container with '--user=root' (risky)"
  exit 1
fi
printf "Unbound version: $(unbound -h | grep "Version" | cut -d" " -f2)\n"
printf "Unbound listening UDP port: $LISTENINGPORT\n"
sed -i "s/port: 1053/port: $LISTENINGPORT/" /etc/unbound/unbound.conf
printf "Verbosity level set to $VERBOSITY\n"
sed -i "s/verbosity: 0/verbosity: $VERBOSITY/" /etc/unbound/unbound.conf
printf "Verbosity details level set to $VERBOSITY_DETAILS\n"
[ $VERBOSITY_DETAILS = 0 ] || ARGS=-$(for i in `seq $VERBOSITY_DETAILS`; do printf "v"; done)
printf "Malicious hostnames and ips blocking is $BLOCK_MALICIOUS\n"
if [ "$BLOCK_MALICIOUS" = "on" ] && [ ! -f /etc/unbound/blocks-malicious.conf ]; then
    printf "Extracting malicious hostnames archive...\n"
    tar -xjf /etc/unbound/malicious-hostnames.bz2 -C /etc/unbound/
    printf "Extracting malicious IPs archive...\n"
    tar -xjf /etc/unbound/malicious-ips.bz2 -C /etc/unbound/
    printf "Building blocks-malicious.conf for Unbound...\n"
    while read hostname; do
        echo "local-zone: \""$hostname"\" static" >> /etc/unbound/blocks-malicious.conf
    done < /etc/unbound/malicious-hostnames
    while read ip; do
        echo "private-address: $ip" >> /etc/unbound/blocks-malicious.conf
    done < /etc/unbound/malicious-ips
    printf " => $(cat /etc/unbound/malicious-hostnames | wc -l ) malicious hostnames and $(cat /etc/unbound/malicious-ips | wc -l) malicious IP addresses added\n"
    rm -f /etc/unbound/malicious-hostnames* /etc/unbound/malicious-ips*
else
    touch /etc/unbound/blocks-malicious.conf
fi
unbound -d $ARGS
status=$?
printf "\n =========================================\n"
printf " Unbound exit with status $status\n"
printf " =========================================\n"