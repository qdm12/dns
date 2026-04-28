package dnssec

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func Test_desiredZoneToZoneNames(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		desiredZone string
		zoneNames   []string
	}{
		"root": {
			desiredZone: ".",
			zoneNames:   []string{"."},
		},
		"com": {
			desiredZone: "com.",
			zoneNames:   []string{".", "com."},
		},
		"example.com": {
			desiredZone: "example.com.",
			zoneNames:   []string{".", "com.", "example.com."},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			zoneNames := desiredZoneToZoneNames(testCase.desiredZone)
			assert.Equal(t, testCase.zoneNames, zoneNames)
		})
	}
}

type staticResponseHandler struct {
	response *dns.Msg
}

func (h staticResponseHandler) ServeDNS(w dns.ResponseWriter, _ *dns.Msg) {
	_ = w.WriteMsg(h.response.Copy())
}

func Test_queryDS(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		response   *dns.Msg
		noData     bool
		nxDomain   bool
		errWrapped error
	}{
		"unsigned_nodata_is_treated_as_insecure": {
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{Rcode: dns.RcodeSuccess},
			},
			noData: true,
		},
		"unsigned_positive_answer_is_rejected": {
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{Rcode: dns.RcodeSuccess},
				Answer: []dns.RR{&dns.DS{Hdr: dns.RR_Header{
					Name:   "example.com.",
					Rrtype: dns.TypeDS,
					Class:  dns.ClassINET,
				}}},
			},
			errWrapped: errDSAndNSECAbsent,
		},
		"unsigned_nxdomain_is_treated_as_insecure": {
			response: &dns.Msg{
				MsgHdr: dns.MsgHdr{Rcode: dns.RcodeNameError},
			},
			nxDomain: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			handler := staticResponseHandler{response: testCase.response}
			response, err := queryDS(handler, "example.com.", dns.ClassINET)

			assert.ErrorIs(t, err, testCase.errWrapped)
			if testCase.noData {
				assert.True(t, response.isNoData())
			}
			if testCase.nxDomain {
				assert.True(t, response.isNXDomain())
			}
		})
	}
}
