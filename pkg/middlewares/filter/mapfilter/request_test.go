package mapfilter

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func Test_FilterRequest(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		makeFilter func(ctrl *gomock.Controller) *Filter
		request    *dns.Msg
		blocked    bool
	}{
		"no_question": {
			makeFilter: func(_ *gomock.Controller) *Filter {
				return &Filter{}
			},
			request: &dns.Msg{},
		},
		"no_filter": {
			makeFilter: func(_ *gomock.Controller) *Filter {
				return &Filter{}
			},
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "host.domain.com.", Qclass: dns.ClassINET, Qtype: dns.TypeA},
				},
			},
		},
		"no_matching_filters": {
			makeFilter: func(_ *gomock.Controller) *Filter {
				return &Filter{
					fqdnHostnames: map[string]struct{}{
						"example.org.": {},
					},
				}
			},
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "host.domain.com.", Qclass: dns.ClassINET, Qtype: dns.TypeA},
				},
			},
		},
		"blocked_exact_match": {
			makeFilter: func(ctrl *gomock.Controller) *Filter {
				metrics := NewMockMetrics(ctrl)
				metrics.EXPECT().HostnamesFilteredInc("IN", "A")
				return &Filter{
					fqdnHostnames: map[string]struct{}{
						"host.domain.com.": {},
					},
					metrics: metrics,
				}
			},
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "host.domain.com.", Qclass: dns.ClassINET, Qtype: dns.TypeA},
				},
			},
			blocked: true,
		},
		"blocked_parent_match": {
			makeFilter: func(ctrl *gomock.Controller) *Filter {
				metrics := NewMockMetrics(ctrl)
				metrics.EXPECT().HostnamesFilteredInc("IN", "A")
				return &Filter{
					fqdnHostnames: map[string]struct{}{
						"domain.com.": {},
					},
					metrics: metrics,
				}
			},
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "host.domain.com.", Qclass: dns.ClassINET, Qtype: dns.TypeA},
				},
			},
			blocked: true,
		},
		"blocked_grand_parent_match": {
			makeFilter: func(ctrl *gomock.Controller) *Filter {
				metrics := NewMockMetrics(ctrl)
				metrics.EXPECT().HostnamesFilteredInc("IN", "A")
				return &Filter{
					fqdnHostnames: map[string]struct{}{
						"domain.com.": {},
					},
					metrics: metrics,
				}
			},
			request: &dns.Msg{
				Question: []dns.Question{
					{Name: "xyz.host.domain.com.", Qclass: dns.ClassINET, Qtype: dns.TypeA},
				},
			},
			blocked: true,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			filter := testCase.makeFilter(ctrl)

			blocked := filter.FilterRequest(testCase.request)
			assert.Equal(t, testCase.blocked, blocked)
		})
	}
}
