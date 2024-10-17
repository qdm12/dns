package doh

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"net/url"
	"time"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/pkg/provider"
)

// hardcodedConn is a net.Conn implementation that is used
// to fake a DNS server connection and respond with hardcoded
// IP address values for specific FQDNs. It is notably used
// to resolve DoH server names to IP addresses for the DoH
// HTTP client.
type hardcodedConn struct {
	// Injected fields
	fqdnToIPv4 map[string][]netip.Addr
	fqdnToIPv6 map[string][]netip.Addr

	// Internal fields
	dataForClient []byte
	finished      bool
}

// Write is called by the DNS client to send its DNS request,
// and it computes the response data bytes from hardcoded data
// to be sent back to the client when it calls the Read method.
func (c *hardcodedConn) Write(b []byte) (n int, err error) {
	if c.finished {
		return 0, fmt.Errorf("writing to hardcoded connection: %w", io.ErrClosedPipe)
	}

	request := new(dns.Msg)
	const lengthPrefixSize = 2 // we have no use of the total length
	err = request.Unpack(b[lengthPrefixSize:])
	if err != nil {
		return 0, fmt.Errorf("unpacking DNS request message: %w", err)
	}

	response, err := c.buildResponse(request)
	if err != nil {
		return 0, fmt.Errorf("building response: %w", err)
	}

	packedResponse, err := response.Pack()
	if err != nil {
		return 0, fmt.Errorf("packing response: %w", err)
	}
	const maxPackedResponseLength = 65535
	if len(packedResponse) > maxPackedResponseLength {
		panic("packed response is bigger than 65535 bytes")
	}
	packedLength := uint16(len(packedResponse)) //nolint:gosec

	c.dataForClient = make([]byte, lengthPrefixSize, lengthPrefixSize+len(packedResponse))
	binary.BigEndian.PutUint16(c.dataForClient, packedLength)
	c.dataForClient = append(c.dataForClient, packedResponse...)

	return len(b), nil
}

var (
	errHardcodedNoIPFound                = errors.New("no IP address found")
	errHardcodedQuestionTypeNotSupported = errors.New("question type not supported")
)

func (c *hardcodedConn) buildResponse(request *dns.Msg) (response *dns.Msg, err error) {
	response = new(dns.Msg)
	response = response.SetReply(request)

	// Track names for which we found at least one IP address,
	// whether IPv4 or IPv6, across all questions received.
	// Any name with no matching IP address found makes the
	// function return an error indicating the issue, since
	// this hardcoded resolver should exclusively be used to
	// resolve DoH server names to IP addresses.
	nameToIPFound := map[string]bool{}

	for _, question := range request.Question {
		answers, err := questionToHardcodedAnswers(question,
			c.fqdnToIPv4, c.fqdnToIPv6, nameToIPFound)
		if err != nil {
			return nil, err
		}
		response.Answer = append(response.Answer, answers...)
	}

	for name, found := range nameToIPFound {
		if found {
			continue
		}
		return nil, fmt.Errorf("%w: for name: %s", errHardcodedNoIPFound, name)
	}

	return response, nil
}

func questionToHardcodedAnswers(question dns.Question,
	fqdnToIPv4, fqdnToIPv6 map[string][]netip.Addr,
	nameToIPFound map[string]bool,
) (answers []dns.RR, err error) {
	_, exists := nameToIPFound[question.Name]
	if !exists {
		nameToIPFound[question.Name] = false
	}

	const hardcodedTTL = 3600 * 24 // TTL of 1 day
	switch question.Qtype {
	case dns.TypeA:
		ips := fqdnToIPv4[question.Name]
		if len(ips) == 0 {
			return nil, nil
		}
		nameToIPFound[question.Name] = true
		header := dns.RR_Header{
			Name:   question.Name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    hardcodedTTL,
		}
		answers = make([]dns.RR, len(ips))
		for i, ip := range ips {
			answers[i] = &dns.A{
				Hdr: header,
				A:   ip.AsSlice(),
			}
		}
		return answers, nil
	case dns.TypeAAAA:
		ips := fqdnToIPv6[question.Name]
		if len(ips) == 0 {
			return nil, nil
		}
		nameToIPFound[question.Name] = true
		header := dns.RR_Header{
			Name:   question.Name,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    hardcodedTTL,
		}
		answers = make([]dns.RR, len(ips))
		for i, ip := range ips {
			answers[i] = &dns.AAAA{
				Hdr:  header,
				AAAA: ip.AsSlice(),
			}
		}
		return answers, nil
	default:
		return nil, fmt.Errorf("%w: %s", errHardcodedQuestionTypeNotSupported,
			dns.TypeToString[question.Qtype])
	}
}

// Read is called by the DNS client to read the DNS response,
// and simply writes to b the pre-computed response data bytes.
func (c *hardcodedConn) Read(b []byte) (n int, err error) {
	if c.finished {
		return 0, fmt.Errorf("reading from hardcoded connection: %w", io.EOF)
	}

	n = copy(b, c.dataForClient)
	c.dataForClient = c.dataForClient[n:]
	c.finished = len(c.dataForClient) == 0
	return n, nil
}

func (c *hardcodedConn) Close() error {
	c.finished = true
	return nil
}

func (c *hardcodedConn) LocalAddr() net.Addr {
	return nil
}

func (c *hardcodedConn) RemoteAddr() net.Addr {
	return nil
}

func (c *hardcodedConn) SetDeadline(time.Time) error {
	return nil
}

func (c *hardcodedConn) SetReadDeadline(time.Time) error {
	return nil
}

func (c *hardcodedConn) SetWriteDeadline(time.Time) error {
	return nil
}

func dohServersToHardcodedMaps(dohServers []provider.DoHServer, ipVersion string) (
	fqdnToIPv4, fqdnToIPv6 map[string][]netip.Addr,
) {
	fqdnToIPv4 = make(map[string][]netip.Addr, len(dohServers))
	fqdnToIPv6 = make(map[string][]netip.Addr, len(dohServers))
	for _, dohServer := range dohServers {
		u, err := url.Parse(dohServer.URL)
		if err != nil {
			// url should be valid at this point,
			// so we panic if there is a parse error.
			panic(err)
		}
		fqdn := dns.Fqdn(u.Hostname())

		ips := make([]netip.Addr, len(dohServer.IPv4))
		copy(ips, dohServer.IPv4)
		fqdnToIPv4[fqdn] = ips

		if ipVersion == "ipv6" {
			ips := make([]netip.Addr, len(dohServer.IPv6))
			copy(ips, dohServer.IPv6)
			fqdnToIPv6[fqdn] = ips
		}
	}
	return fqdnToIPv4, fqdnToIPv6
}
