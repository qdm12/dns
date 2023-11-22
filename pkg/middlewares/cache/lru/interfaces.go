package lru

type Metrics interface { //nolint:interfacebloat
	SetCacheType(cacheType string)
	CacheInsertInc()
	CacheRemoveInc()
	CacheMoveInc()
	CacheGetEmptyInc()
	CacheInsertEmptyInc()
	CacheRemoveEmptyInc()
	CacheHitInc()
	CacheMissInc()
	CacheExpiredInc()
	CacheMaxEntriesSet(maxEntries int)
}
