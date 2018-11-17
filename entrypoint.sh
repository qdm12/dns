#!/bin/sh

printf " =========================================\n"
printf " =========================================\n"
printf " === CLOUDFLARE DNS OVER TLS CONTAINER ===\n"
printf " =========================================\n"
printf " =========================================\n"
printf " == by github.com/qdm12 - Quentin McGaw ==\n\n"

user=$(whoami)
printf "Running as $user\n"
[ "$user" != "nonrootuser" ] || sed -i 's/username: "nonrootuser"/username: ""/' /etc/unbound/unbound.conf
[ -f /etc/unbound/include.conf ] || touch /etc/unbound/include.conf
test -r /etc/unbound/include.conf
if [ $? != 0 ]; then
  owneruid=$(stat -c %u /etc/unbound/include.conf)
  ownergid=$(stat -c %g /etc/unbound/include.conf)
  permissions=$(stat -c %a /etc/unbound/include.conf)
  printf "Can't read file include.conf (owner UID $owneruid, owner GID $ownergid, permissions $permissions)\n"
  if [ "$user" = "nonrootuser" ]; then
    printf "1) Make sure include.conf has read permission (chmod 400 include.conf)\n"
    printf "2) Either:\n"
    printf "   a) Run the container with '--user=$owneruid'\n"
    printf "   b) Change include.conf ownership 'chown 1000:1000 include.conf'\n"
  fi
  exit 1
fi
VERBOSITY=${VERBOSITY:-1}
VERBOSITY_DETAILS=${VERBOSITY_DETAILS:-0}
LISTENINGPORT=${LISTENINGPORT:-53}
BLOCK_MALICIOUS=${BLOCK_MALICIOUS:-on}
if [ -z $(echo $VERBOSITY | grep -E '^[0-9]+$') ] || [ $VERBOSITY -gt 5 ]; then
  printf "Environment variable VERBOSITY=$VERBOSITY must be a positive integer between 0 and 5\n"
  exit 1
fi
if [ -z $(echo $VERBOSITY_DETAILS | grep -E '^[0-9]+$') ] || [ $VERBOSITY_DETAILS -gt 4 ]; then
  printf "Environment variable VERBOSITY_DETAILS=$VERBOSITY_DETAILS must be a positive integer between 0 and 4\n"
  exit 1
fi
if [ "$BLOCK_MALICIOUS" != "on" ] && [ "$BLOCK_MALICIOUS" != "off" ]; then
  printf "Environment variable BLOCK_MALICIOUS=$BLOCK_MALICIOUS must be 'on' or 'off'\n"
  exit 1
fi
if [ -z $(echo $LISTENINGPORT | grep -E '^[0-9]+$') ] || [ $LISTENINGPORT -lt 1 ] || [ $LISTENINGPORT -gt 65535 ]; then
  printf "Environment variable LISTENINGPORT=$LISTENINGPORT must be a positive integer between 1 and 65535\n"
  exit 1
fi
printf "Unbound version: $(unbound -h | grep "Version" | cut -d" " -f2)\n"
printf "Unbound listening UDP port: $LISTENINGPORT\n"
sed -i "s/port: 53/port: $LISTENINGPORT/" /etc/unbound/unbound.conf
printf "Verbosity level set to $VERBOSITY on 5\n"
sed -i "s/verbosity: 0/verbosity: $VERBOSITY/" /etc/unbound/unbound.conf
printf "Verbosity details level set to $VERBOSITY_DETAILS on 4\n"
[ $VERBOSITY_DETAILS = 0 ] || ARGS=-$(for i in `seq $VERBOSITY_DETAILS`; do printf "v"; done)
printf "Malicious hostnames and ips blocking is $BLOCK_MALICIOUS\n"
if [ "$BLOCK_MALICIOUS" = "on" ]; then
  tar -xjf /etc/unbound/blocks-malicious.bz2 -C /etc/unbound/
  printf "$(cat /etc/unbound/blocks-malicious.conf | grep "local-zone" | wc -l ) malicious hostnames and $(cat /etc/unbound/blocks-malicious.conf | grep "private-address" | wc -l) malicious IP addresses blacklisted\n"
else
  echo "" > /etc/unbound/blocks-malicious.conf
fi
unbound -d $ARGS
status=$?
printf "\n =========================================\n"
printf " Unbound exit with status $status\n"
printf " =========================================\n"
