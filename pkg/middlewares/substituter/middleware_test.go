package substituter

import (
	net "net"
	"net/netip"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	t.Parallel()

	settings := Settings{
		Substitutions: []Substitution{
			{Name: "github.com", IPs: []netip.Addr{netip.MustParseAddr("1.2.3.4")}},
		},
	}

	middleware, err := New(settings)
	require.NoError(t, err)

	expectedMiddleware := &Middleware{
		mapping: map[questionKey][]dns.RR{
			{
				Name:   "github.com.",
				Qtype:  dns.TypeA,
				Qclass: dns.ClassINET,
			}: {&dns.A{
				Hdr: dns.RR_Header{
					Name:   "github.com.",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				A: net.IP{1, 2, 3, 4},
			}},
		},
	}
	assert.Equal(t, expectedMiddleware, middleware)

	next := dns.HandlerFunc(func(_ dns.ResponseWriter, _ *dns.Msg) {})
	handler := middleware.Wrap(next)

	request := &dns.Msg{Question: []dns.Question{
		{Name: "github.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
	}}

	ctrl := gomock.NewController(t)
	writer := NewMockResponseWriter(ctrl)
	substitutedResponse := &dns.Msg{
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   "github.com.",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				A: net.IP{1, 2, 3, 4},
			},
		},
	}
	substitutedResponse.SetReply(request)

	writer.EXPECT().WriteMsg(substitutedResponse)

	handler.ServeDNS(writer, request)

	err = middleware.Stop()
	require.NoError(t, err)
}

func Test_handler_ServeDNS(t *testing.T) {
	t.Parallel()

	request := &dns.Msg{
		Question: []dns.Question{
			{Name: "github.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
		},
	}
	response := &dns.Msg{
		Answer: []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   "github.com.",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				A: net.IP{1, 2, 3, 4},
			},
		},
	}
	response.SetReply(request)

	testCases := map[string]struct {
		settings              Settings
		responseWriterBuilder func(ctrl *gomock.Controller) dns.ResponseWriter
	}{
		"no_substitution": {
			responseWriterBuilder: func(ctrl *gomock.Controller) dns.ResponseWriter {
				return NewMockResponseWriter(ctrl)
			},
		},
		"substitution": {
			settings: Settings{
				Substitutions: []Substitution{
					{Name: "github.com", IPs: []netip.Addr{netip.MustParseAddr("1.2.3.4")}},
				},
			},
			responseWriterBuilder: func(ctrl *gomock.Controller) dns.ResponseWriter {
				writer := NewMockResponseWriter(ctrl)
				writer.EXPECT().WriteMsg(response).Return(nil)
				return writer
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			middleware, err := New(testCase.settings)
			require.NoError(t, err)

			next := dns.HandlerFunc(func(_ dns.ResponseWriter, m *dns.Msg) {
				assert.Equal(t, request, m)
			})
			handler := middleware.Wrap(next)

			responseWriter := testCase.responseWriterBuilder(ctrl)

			handler.ServeDNS(responseWriter, request)
		})
	}
}
