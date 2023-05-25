// Package noop defines a No-op metric implementation for the cache.
package noop

type Metrics struct{}

func New() (metrics *Metrics) {
	return new(Metrics)
}

func (m *Metrics) SetCacheType(string)    {}
func (m *Metrics) CacheInsertInc()        {}
func (m *Metrics) CacheRemoveInc()        {}
func (m *Metrics) CacheMoveInc()          {}
func (m *Metrics) CacheGetEmptyInc()      {}
func (m *Metrics) CacheInsertEmptyInc()   {}
func (m *Metrics) CacheHitInc()           {}
func (m *Metrics) CacheMissInc()          {}
func (m *Metrics) CacheExpiredInc()       {}
func (m *Metrics) CacheMaxEntriesSet(int) {}
