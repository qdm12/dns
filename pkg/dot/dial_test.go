package dot

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_settingsToServers(t *testing.T) {
	t.Parallel()

	settings := ResolverSettings{
		DoTProviders: []string{
			"cloudflare", "google",
		},
		DNSProviders: []string{
			"cira family",
		},
	}

	dotServers, dnsServers, err := settingsToServers(settings)
	require.NoError(t, err)

	assert.Equal(t, []provider.DoTServer{
		provider.Cloudflare().DoT,
		provider.Google().DoT,
	}, dotServers)
	assert.Equal(t, []provider.DNSServer{
		provider.CiraFamily().DNS,
	}, dnsServers)
}

//go:generate mockgen -destination=mock_picker_test.go -package $GOPACKAGE -mock_names Interface=MockPicker github.com/qdm12/dns/v2/internal/picker Interface
//go:generate mockgen -destination=mock_dot_metrics_test.go -package $GOPACKAGE -mock_names Interface=MockDoTMetrics github.com/qdm12/dns/v2/pkg/dot/metrics Interface
//go:generate mockgen -destination=mock_warner_test.go -package $GOPACKAGE github.com/qdm12/dns/v2/pkg/log Warner

func Test_pickNameAddress(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	picker := NewMockPicker(ctrl)
	servers := []provider.DoTServer{
		provider.Cloudflare().DoT,
		provider.Google().DoT,
	}
	const ipv6 = true

	picker.EXPECT().DoTServer(servers).Return(servers[0])
	picker.EXPECT().DoTIP(servers[0], ipv6).Return(servers[0].IPv6[0])

	name, address := pickNameAddress(picker, servers, ipv6)

	assert.Equal(t, "cloudflare-dns.com", name)
	assert.Equal(t, "[2606:4700:4700::1111]:853", address)
}

func Test_dialPlaintext(t *testing.T) {
	t.Parallel()

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	testCases := map[string]struct {
		ctx           context.Context
		ipv6          bool
		dnsServers    []provider.DNSServer
		expectedAddr  string
		metricOutcome string
		err           error
	}{
		"success": {
			ctx: context.Background(),
			dnsServers: []provider.DNSServer{
				provider.Cloudflare().DNS,
			},
			expectedAddr:  "1.1.1.1:53",
			metricOutcome: "success",
		},
		"canceled context": {
			ctx: canceledCtx,
			dnsServers: []provider.DNSServer{
				provider.Cloudflare().DNS,
			},
			expectedAddr:  "1.1.1.1:53",
			metricOutcome: "error",
			err:           errors.New("dial udp 1.1.1.1:53: operation was canceled"),
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			warner := NewMockWarner(ctrl)
			if testCase.err != nil {
				warner.EXPECT().Warn(testCase.err.Error())
			}

			metrics := NewMockDoTMetrics(ctrl)
			metrics.EXPECT().DNSDialInc(testCase.expectedAddr, testCase.metricOutcome)

			picker := NewMockPicker(ctrl)
			picker.EXPECT().DNSServer(testCase.dnsServers).
				Return(testCase.dnsServers[0])
			picker.EXPECT().DNSIP(testCase.dnsServers[0], false).
				Return(testCase.dnsServers[0].IPv4[0])

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
