#!/bin/sh

printf " =========================================\n"
printf " =========================================\n"
printf " === CLOUDFLARE DNS OVER TLS CONTAINER ===\n"
printf " =========================================\n"
printf " =========================================\n"
printf " == by github.com/qdm12 - Quentin McGaw ==\n\n"

# Checks parameters and mounted files
user=$(whoami)
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
if [ -z $(echo $VERBOSITY | grep -E '^0|1|2|3|4|5$') ]; then
  printf "Environment variable VERBOSITY=$VERBOSITY must be an integer between 0 and 5\n"
  exit 1
fi
if [ -z $(echo $VERBOSITY_DETAILS | grep -E '^0|1|2|3|4$') ]; then
  printf "Environment variable VERBOSITY_DETAILS=$VERBOSITY_DETAILS must be an integer between 0 and 4\n"
  exit 1
fi
if [ "$BLOCK_MALICIOUS" != "on" ] && [ "$BLOCK_MALICIOUS" != "off" ]; then
  printf "Environment variable BLOCK_MALICIOUS=$BLOCK_MALICIOUS must be 'on' or 'off'\n"
  exit 1
fi
if [ "$BLOCK_NSA" != "on" ] && [ "$BLOCK_NSA" != "off" ]; then
  printf "Environment variable BLOCK_NSA=$BLOCK_NSA must be 'on' or 'off'\n"
  exit 1
fi
if [ -z $(echo $LISTENINGPORT | grep -E '^[0-9]+$') ] || [ $LISTENINGPORT -lt 1 ] || [ $LISTENINGPORT -gt 65535 ]; then
  printf "Environment variable LISTENINGPORT=$LISTENINGPORT must be a positive integer between 1 and 65535\n"
  exit 1
fi

# Modifies configuration according to valid parameters
printf "Running as $user\n"
if [ "$user" != "nonrootuser" ]; then
  sed -i 's/username: .*$/username: "nonrootuser"/' /etc/unbound/unbound.conf
else
  sed -i 's/username: .*$/username: ""/' /etc/unbound/unbound.conf
fi
printf "Unbound version: $(unbound -h | grep "Version" | cut -d" " -f2)\n"
sed -i '/forward-addr/d' /etc/unbound/unbound.conf
case $PROVIDER in
  cloudflare)
    printf "forward-addr: 1.1.1.1@853#cloudflare-dns.com\n" >> /etc/unbound/unbound.conf
    echo "forward-addr: 1.0.0.1@853#cloudflare-dns.com" >> /etc/unbound/unbound.conf
    ;;
  google)
    echo "forward-addr: 8.8.8.8@853#dns.google" >> /etc/unbound/unbound.conf
    echo "forward-addr: 8.8.4.4@853#dns.google" >> /etc/unbound/unbound.conf
    ;;
  quad9)
    echo "forward-addr: 9.9.9.9@853#dns.quad9.net" >> /etc/unbound/unbound.conf
    echo "forward-addr: 149.112.112.112@853#dns.quad9.net" >> /etc/unbound/unbound.conf
    ;;
  quadrant)
    echo "forward-addr: 12.159.2.159@853#dns-tls.qis.io" >> /etc/unbound/unbound.conf
    ;;
  cleanbrowsing)
    echo "forward-addr: 185.228.168.9@853#security-filter-dns.cleanbrowsing.org" >> /etc/unbound/unbound.conf
    echo "forward-addr: 185.228.169.9@853#security-filter-dns.cleanbrowsing.org" >> /etc/unbound/unbound.conf
    ;;
  *)
    printf "Environment variable PROVIDER=$PROVIDER must be 'cloudflare', 'google', 'quad9', 'quadrant' or 'cleanbrowsing'\n"
    exit 1
    ;;
esac
printf "Unbound DNS server: $PROVIDER\n"    
printf "Unbound listening UDP port: $LISTENINGPORT\n"
sed -i "s/port: .*$/port: $LISTENINGPORT/" /etc/unbound/unbound.conf
printf "Verbosity level set to $VERBOSITY on 5\n"
sed -i "s/verbosity: .*$/verbosity: $VERBOSITY/" /etc/unbound/unbound.conf
printf "Verbosity details level set to $VERBOSITY_DETAILS on 4\n"
[ $VERBOSITY_DETAILS = 0 ] || ARGS=-$(for i in `seq $VERBOSITY_DETAILS`; do printf "v"; done)
printf "Malicious hostnames and ips blocking is $BLOCK_MALICIOUS\n"
if [ "$BLOCK_MALICIOUS" = "on" ]; then
  tar -xjf /etc/unbound/blocks-malicious.bz2 -C /etc/unbound/
  printf "$(cat /etc/unbound/blocks-malicious.conf | grep "local-zone" | wc -l ) malicious hostnames and $(cat /etc/unbound/blocks-malicious.conf | grep "private-address" | wc -l) malicious IP addresses blacklisted\n"
else
  echo "" > /etc/unbound/blocks-malicious.conf
fi
printf "NSA hostnames blocking is $BLOCK_NSA\n"
if [ "$BLOCK_NSA" = "on" ]; then
  tar -xjf /etc/unbound/blocks-nsa.bz2 -C /etc/unbound/
  printf "$(cat /etc/unbound/blocks-nsa.conf | grep "local-zone" | wc -l ) NSA hostnames blacklisted\n"
  cat /etc/unbound/blocks-nsa.conf >> /etc/unbound/blocks-malicious.conf
  rm /etc/unbound/blocks-nsa.conf
  sort -u -o /etc/unbound/blocks-malicious.conf /etc/unbound/blocks-malicious.conf
fi
for hostname in ${UNBLOCK//,/ }
do
  printf "Unblocking hostname $hostname\n"
  sed -i "/$hostname/d" /etc/unbound/blocks-malicious.conf
done
unbound -d $ARGS
status=$?
printf "\n =========================================\n"
printf " Unbound exit with status $status\n"
printf " =========================================\n"
