# DNSSEC validator

This package implements a DNSSEC validator middleware for a DNS forwarding server.
It performs all queries through the next DNS handler middleware.... TODO

## Comments

Comments are aimed to be minimal and clear code is preferred over comments.
However, there are a few references to specific sections of IETF RFCs especially when it comes to function comments.

## Terminology used

Teminology used in the code aims at being as close as possible to IETF RFCs.

When this is unclear, there are a few specific rules used, for example:

- Using `validate` or `verify`, see the `validate vs. verify` in [RFC4949](https://datatracker.ietf.org/doc/html/rfc4949)

## Documentation used

### IETF RFCs

- [RFC4033](https://datatracker.ietf.org/doc/html/rfc4033)
- [RFC5155 on NSEC3](https://datatracker.ietf.org/doc/html/rfc5155)
- [RFC8624 on DNSKEY algorithms](https://datatracker.ietf.org/doc/html/rfc8624#section-3.1)

### Blog posts

<https://blog.nlnetlabs.nl/the-peculiar-case-of-nsec-processing-using-expanded-wildcard-records/>
<https://wander.science/projects/dns/dnssec-resolver-test/>

### Videos

- [DNSSEC Series #5 Record types, keys, signatures and NSEC](https://www.youtube.com/watch?v=FGs9kbdgMXE&t=2825s)

## Tools used to help debugging

- [DNSViz](https://dnsviz.net/)
- [DNSSEC Analyzer](https://dnssec-analyzer.verisignlabs.com)
