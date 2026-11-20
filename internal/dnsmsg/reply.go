// Package dnsmsg contains helpers around the github.com/miekg/dns
// message type.
package dnsmsg

import (
	"github.com/miekg/dns"
)

// SetReply sets the response header fields derived from the request
// (id, response bit, opcode, RD and CD bits, and question), and
// preserves the response code of the response.
//
// This is required because dns.Msg.SetReply unconditionally sets the
// rcode to NOERROR (RcodeSuccess). Calling it on a response received
// from an upstream resolver rewrites error rcodes such as NXDOMAIN,
// SERVFAIL and REFUSED into NOERROR with 0 answers (NODATA).
// This breaks musl libc resolvers (Alpine images): when a name is
// suffixed with a search domain, glibc moves on to the next search
// domain after a NODATA answer, whereas musl considers NODATA to be
// final and fails the lookup. See
// https://github.com/passteque/gluetun/issues/3552
//
// To build a response with an explicit rcode, use dns.Msg.SetRcode
// instead, which sets the reply fields and the rcode at once.
func SetReply(response, request *dns.Msg) *dns.Msg {
	rcode := response.Rcode
	response.SetReply(request)
	response.Rcode = rcode
	return response
}
