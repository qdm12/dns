package mapfilter

import (
	"fmt"
	"net/netip"
	"sync"
)

type Filter struct {
	fqdnHostnames     map[string]struct{}
	ipv4              map[[4]byte]struct{}
	ipv6              map[[16]byte]struct{}
	ipPrefixes        []netip.Prefix
	privateIPPrefixes []netip.Prefix
	allowRebindNames  map[string]struct{}
	metrics           Metrics
	logger            Logger
	updateLock        sync.RWMutex
}

func New(settings Settings) (filter *Filter, err error) {
	settings.SetDefaults()

	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("settings validation: %w", err)
	}

	filter = &Filter{
		privateIPPrefixes: getPrivateIPPrefixes(),
		metrics:           settings.Metrics,
		logger:            settings.Logger,
	}

	err = filter.Update(settings.Update)
	if err != nil {
		return nil, err
	}

	return filter, nil
}
