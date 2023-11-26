package mapfilter

import (
	"fmt"
	"net/netip"
	"sync"
)

type Filter struct {
	fqdnHostnames     map[string]struct{}
	ips               map[netip.Addr]struct{}
	ipPrefixes        []netip.Prefix
	privateIPPrefixes []netip.Prefix
	metrics           Metrics
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
	}

	err = filter.Update(settings.Update)
	if err != nil {
		return nil, err
	}

	return filter, nil
}
