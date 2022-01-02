package dot

import (
	"context"
	"fmt"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/server"
)

func newDNSHandler(ctx context.Context, settings ServerSettings) (
	h dns.Handler, err error) {
	dial, err := newDoTDial(settings.Resolver)
	if err != nil {
		return nil, fmt.Errorf("cannot create DoT dial: %w", err)
	}

	exchange := server.NewExchange("DoT", dial, settings.Logger)

	return server.New(ctx, exchange, settings.Filter,
		settings.Cache, settings.Logger), nil
}
