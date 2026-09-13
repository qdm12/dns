package dnssec

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/miekg/dns"
	internaldnssec "github.com/qdm12/dns/v2/internal/dnssec"
)

// Middleware implements a DNSSEC validator.
type Middleware struct {
	settings        Settings
	wrapping        atomic.Bool
	validator       *internaldnssec.Validator
	stopRefreshLoop context.CancelFunc
}

func New(settings Settings) (middleware *Middleware, err error) {
	settings.SetDefaults()

	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	return &Middleware{
		settings:  settings,
		validator: internaldnssec.New(),
	}, nil
}

func (m *Middleware) String() string {
	return "DNSSEC validator"
}

// Wrap wraps the DNS handler with the middleware.
func (m *Middleware) Wrap(next dns.Handler) dns.Handler { //nolint:ireturn
	previousWrapping := m.wrapping.Swap(true)
	if previousWrapping {
		panic("DNSSEC middleware cannot wrap more than once")
	}

	handler := newHandler(m.settings.Logger, m.validator, next)
	if *m.settings.RootTrustAnchorRefreshPeriod > 0 {
		m.startRefreshLoop(next)
	}
	return handler
}

func (m *Middleware) Stop() (err error) {
	if m.stopRefreshLoop != nil {
		m.stopRefreshLoop()
		m.stopRefreshLoop = nil
	}
	return nil
}

func (m *Middleware) startRefreshLoop(next dns.Handler) {
	ctx, cancel := context.WithCancel(context.Background())
	m.stopRefreshLoop = cancel

	go func() {
		ticker := time.NewTicker(*m.settings.RootTrustAnchorRefreshPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := m.validator.RefreshRootTrustAnchors(next)
				if err != nil {
					m.settings.Logger.Warn("refreshing root trust anchors: " + err.Error())
					continue
				}
				m.settings.Logger.Info("refreshed root trust anchors")
			}
		}
	}()
}
