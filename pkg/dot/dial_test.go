package dot

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/qdm12/dns/internal/mockhelp"
	"github.com/qdm12/dns/pkg/dot/metrics/mock_metrics"
	"github.com/qdm12/dns/pkg/log/mock_log"
	"github.com/qdm12/dns/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_settingsToServers(t *testing.T) {
	t.Parallel()

	settings := ResolverSettings{
		DoTProviders: []provider.Provider{
			provider.Cloudflare(),
			provider.Google(),
		},
		DNSProviders: []provider.Provider{
			provider.CiraFamily(),
		},
	}

	dotServers, dnsServers := settingsToServers(settings)

	assert.Equal(t, []provider.DoTServer{
		provider.Cloudflare().DoT(),
		provider.Google().DoT(),
	}, dotServers)
	assert.Equal(t, []provider.DNSServer{
		provider.CiraFamily().DNS(),
	}, dnsServers)
}

func Test_pickNameAddress(t *testing.T) {
	t.Parallel()

	picker := newPicker()
	servers := []provider.DoTServer{
		provider.Cloudflare().DoT(),
		provider.Google().DoT(),
	}
	const ipv6 = true

	const tries = 10

	for i := 0; i < tries; i++ {
		name, address := pickNameAddress(picker, servers, ipv6)

		switch name {
		case "dns.google":
			switch address {
			case "[2001:4860:4860::8844]:853", "[2001:4860:4860::8888]:853":
			default:
				t.Errorf("unexpected address for dns.google: %s", address)
			}
		case "cloudflare-dns.com":
			switch address {
			case "[2606:4700:4700::1111]:853", "[2606:4700:4700::1001]:853":
			default:
				t.Errorf("unexpected address for cloudflare-dns.com: %s", address)
			}
		default:
			t.Errorf("unexpected name: %s", name)
		}
	}
}

func Test_dialPlaintext(t *testing.T) {
	t.Parallel()

	picker := newPicker()

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	testCases := map[string]struct {
		ctx           context.Context
		ipv6          bool
		dnsServers    []provider.DNSServer
		expectedAddrs []string
		metricOutcome string
		err           error
	}{
		"success": {
			ctx: context.Background(),
			dnsServers: []provider.DNSServer{
				provider.Cloudflare().DNS(),
			},
			expectedAddrs: []string{"1.1.1.1:53", "1.0.0.1:53"},
			metricOutcome: "success",
		},
		"canceled context": {
			ctx: canceledCtx,
			dnsServers: []provider.DNSServer{
				{IPv4: []net.IP{net.IPv4(1, 1, 1, 1)}},
			},
			expectedAddrs: []string{"1.1.1.1:53"},
			metricOutcome: "error",
			err:           errors.New("dial udp 1.1.1.1:53: operation was canceled"),
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			warner := mock_log.NewMockWarner(ctrl)
			if testCase.err != nil {
				warner.EXPECT().Warn(testCase.err.Error())
			}

			metrics := mock_metrics.NewMockInterface(ctrl)
			metrics.EXPECT().DNSDialInc(
				mockhelp.NewMatcherOneOf(testCase.expectedAddrs...),
				testCase.metricOutcome)

			dialer := &net.Dialer{} // cannot mock

			conn, err := dialPlaintext(testCase.ctx, dialer, picker,
				testCase.ipv6, testCase.dnsServers, warner, metrics)

			if conn != nil {
				err := conn.Close()
				require.NoError(t, err)
			}

			if testCase.err != nil {
				require.Error(t, err)
				assert.Equal(t, testCase.err.Error(), err.Error())
				assert.Nil(t, conn)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, &net.UDPConn{}, conn)
			}
		})
	}
}
