package substituter

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"

	"github.com/miekg/dns"
)

type Substitution struct {
	Name  string       `json:"name"`
	Type  string       `json:"type"`
	Class string       `json:"class"`
	TTL   uint32       `json:"ttl"`
	IPs   []netip.Addr `json:"ips"`
}

func (s *Substitution) setDefaults() {
	if s.Type == "" {
		defaultType := dns.TypeA // no IP, or IP is no valid or IP is IPv4
		if len(s.IPs) > 0 && s.IPs[0].Is6() {
			defaultType = dns.TypeAAAA
		}
		s.Type = dns.TypeToString[defaultType]
	}
	if s.Class == "" {
		s.Class = dns.ClassToString[dns.ClassINET]
	}
	if s.TTL == 0 {
		const defaultTTL = 300
		s.TTL = defaultTTL
	}
}

var (
	ErrNameIsEmpty        = errors.New("name is empty")
	ErrTypeIsUnknown      = errors.New("type is unknown")
	ErrTypeIsNotSupported = errors.New("type is not supported")
	ErrClassIsUnknown     = errors.New("class is unknown")
	ErrIPVersionMix       = errors.New("IP version mix")
)

func (s *Substitution) validate() (err error) {
	if s.Name == "" {
		return fmt.Errorf("%w", ErrNameIsEmpty)
	}

	qType, ok := dns.StringToType[strings.ToUpper(s.Type)]
	if !ok {
		return fmt.Errorf("%w: %s", ErrTypeIsUnknown, s.Type)
	}
	switch qType {
	case dns.TypeA, dns.TypeAAAA:
		if hasIPVersionMix(s.IPs) {
			return fmt.Errorf("%w: for substitution %s", ErrIPVersionMix, s)
		}
	default:
		return fmt.Errorf("%w: %s", ErrTypeIsNotSupported, s.Type)
	}

	_, ok = dns.StringToClass[strings.ToUpper(s.Class)]
	if !ok {
		return fmt.Errorf("%w: %s", ErrClassIsUnknown, s.Class)
	}

	return nil
}

func hasIPVersionMix(ips []netip.Addr) bool {
	var hasIPv4, hasIPv6 bool
	for _, ip := range ips {
		if ip.Is4() {
			hasIPv4 = true
		} else if ip.Is6() {
			hasIPv6 = true
		}
		if hasIPv4 && hasIPv6 {
			return true
		}
	}
	return false
}

func (s Substitution) String() string {
	var rr string
	qType := dns.StringToType[strings.ToUpper(s.Type)]
	switch qType {
	case dns.TypeA, dns.TypeAAAA:
		ipStrings := make([]string, len(s.IPs))
		for i, ip := range s.IPs {
			ipStrings[i] = ip.String()
		}
		rr = strings.Join(ipStrings, ", ")
	default:
		panic("unsupported type")
	}
	return fmt.Sprintf("%s %s %s -> %s with ttl %d",
		s.Name, s.Type, s.Class, rr, s.TTL)
}

func (s *Substitution) toQuestion() dns.Question {
	return dns.Question{
		Name:   dns.Fqdn(s.Name),
		Qtype:  dns.StringToType[strings.ToUpper(s.Type)],
		Qclass: dns.StringToClass[strings.ToUpper(s.Class)],
	}
}

func (s *Substitution) toRRs() (rrs []dns.RR) {
	header := dns.RR_Header{
		Name:   dns.Fqdn(s.Name),
		Rrtype: dns.StringToType[strings.ToUpper(s.Type)],
		Class:  dns.StringToClass[strings.ToUpper(s.Class)],
		Ttl:    s.TTL,
	}
	switch header.Rrtype {
	case dns.TypeA:
		if len(s.IPs) == 0 {
			return []dns.RR{&dns.A{Hdr: header}}
		}
		rrs = make([]dns.RR, len(s.IPs))
		for i, ip := range s.IPs {
			rrs[i] = &dns.A{
				Hdr: header,
				A:   ip.AsSlice(),
			}
		}
	case dns.TypeAAAA:
		if len(s.IPs) == 0 {
			return []dns.RR{&dns.AAAA{Hdr: header}}
		}
		rrs = make([]dns.RR, len(s.IPs))
		for i, ip := range s.IPs {
			rrs[i] = &dns.AAAA{
				Hdr:  header,
				AAAA: ip.AsSlice(),
			}
		}
	default:
		panic("unimplemented")
	}
	return rrs
}
