package mapfilter

import (
	"fmt"
	"net/netip"
	"sync"

	"github.com/qdm12/dns/v2/internal/local"
)

type Filter struct {
	fqdnHostnames      map[string]struct{}
	ipv4               map[[4]byte]struct{}
	ipv6               map[[16]byte]struct{}
	ipPrefixes         []netip.Prefix
	privateIPPrefixes  []netip.Prefix
	allowRebindNames   map[string]struct{}
	allowRebindParents map[string]struct{}
	localChecker       LocalChecker
	metrics            Metrics
	logger             Logger
	updateLock         sync.RWMutex
}

func New(settings Settings) (filter *Filter, err error) {
	settings.SetDefaults()

	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("settings validation: %w", err)
	}

	filter = &Filter{
		privateIPPrefixes: getPrivateIPPrefixes(),
		localChecker:      local.New(settings.PublicNamesAsLocal),
		metrics:           settings.Metrics,
		logger:            settings.Logger,
	}

	err = filter.Update(settings.Update)
	if err != nil {
		return nil, err
	}

	return filter, nil
}
