package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/internal/metrics/prometheus/helpers"
	prom "github.com/qdm12/dns/v2/pkg/metrics/prometheus"
)

type counters struct {
	insert      prometheus.Counter
	move        prometheus.Counter
	remove      prometheus.Counter
	insertEmpty prometheus.Counter
	getEmpty    prometheus.Counter
	deleteEmpty prometheus.Counter
	hit         prometheus.Counter
	expired     prometheus.Counter
	miss        prometheus.Counter
}

func newCounters(settings prom.Settings) (c *counters, err error) {
	prefix := settings.Prefix
	c = &counters{
		insert:      helpers.NewCounter(prefix, "cache_insert", "DNS cache insertions"),
		move:        helpers.NewCounter(prefix, "cache_move", "DNS cache move"),
		remove:      helpers.NewCounter(prefix, "cache_remove", "DNS cache remove"),
		insertEmpty: helpers.NewCounter(prefix, "cache_insertEmpty", "DNS cache insertEmpty"),
		getEmpty:    helpers.NewCounter(prefix, "cache_getEmpty", "DNS cache getEmpty"),
		deleteEmpty: helpers.NewCounter(prefix, "cache_deleteEmpty", "DNS cache deleteEmpty"),
		hit:         helpers.NewCounter(prefix, "cache_hit", "DNS cache hit"),
		expired:     helpers.NewCounter(prefix, "cache_expired", "DNS cache expired"),
		miss:        helpers.NewCounter(prefix, "cache_miss", "DNS cache miss"),
	}

	err = helpers.Register(settings.Registry,
		c.insert, c.move, c.remove, c.insertEmpty, c.getEmpty,
		c.hit, c.expired, c.miss)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *counters) CacheInsertInc()      { c.insert.Inc() }
func (c *counters) CacheMoveInc()        { c.move.Inc() }
func (c *counters) CacheRemoveInc()      { c.remove.Inc() }
func (c *counters) CacheInsertEmptyInc() { c.insertEmpty.Inc() }
func (c *counters) CacheGetEmptyInc()    { c.getEmpty.Inc() }
func (c *counters) CacheRemoveEmptyInc() { c.deleteEmpty.Inc() }
func (c *counters) CacheHitInc()         { c.hit.Inc() }
func (c *counters) CacheExpiredInc()     { c.expired.Inc() }
func (c *counters) CacheMissInc()        { c.miss.Inc() }
