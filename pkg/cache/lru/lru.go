package lru

import (
	"container/list"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/cache/metrics"
)

type LRU struct {
	// Configuration
	maxEntries int

	// State
	kv         map[string]*list.Element
	linkedList *list.List
	mutex      sync.Mutex

	// External objects
	metrics metrics.Interface

	// Mock fields
	timeNow func() time.Time
}

func New(settings Settings) *LRU {
	settings.setDefaults()

	settings.Metrics.SetCacheType(CacheType)
	settings.Metrics.CacheMaxEntriesSet(settings.MaxEntries)

	return &LRU{
		maxEntries: settings.MaxEntries,
		kv:         make(map[string]*list.Element, settings.MaxEntries),
		linkedList: list.New(),
		metrics:    settings.Metrics,
		timeNow:    time.Now,
	}
}

func (l *LRU) Add(request, response *dns.Msg) {
	if isEmpty(request, response) {
		// cannot make key if there is no question
		// and do not store empty response.
		l.metrics.CacheInsertEmptyInc()
		return
	}

	key := makeKey(request)
	expUnix := getExpUnix(response, l.timeNow().Unix())
	responseCopy := response.Copy()

	l.mutex.Lock()
	defer l.mutex.Unlock()

	if listElement, ok := l.kv[key]; ok {
		l.linkedList.MoveToFront(listElement)
		entryPtr := listElement.Value.(*entry)
		entryPtr.expUnix = expUnix
		entryPtr.response = responseCopy
		l.metrics.CacheMoveInc()
		return
	}

	entry := &entry{
		key:      key,
		expUnix:  expUnix,
		response: responseCopy,
	}

	if l.maxEntries > 0 && l.linkedList.Len() == l.maxEntries {
		l.removeOldest()
	}

	listElement := l.linkedList.PushFront(entry)
	l.kv[key] = listElement
	l.metrics.CacheInsertInc()
}

func (l *LRU) Get(request *dns.Msg) (response *dns.Msg) {
	if len(request.Question) == 0 {
		// cannot make key if there is no question
		l.metrics.CacheGetEmptyInc()
		return nil
	}

	key := makeKey(request)
	nowUnix := l.timeNow().Unix()

	l.mutex.Lock()
	defer l.mutex.Unlock()

	listElement, ok := l.kv[key]
	if !ok {
		l.metrics.CacheMissInc()
		return nil
	}

	l.metrics.CacheHitInc()

	l.linkedList.MoveToFront(listElement)
	entryPtr := listElement.Value.(*entry)

	if nowUnix >= entryPtr.expUnix {
		// expired record
		l.remove(listElement)
		l.metrics.CacheExpiredInc()
		return nil
	}

	return entryPtr.response.Copy()
}

// remove removes a list element
// It is NOT thread safe and its parent should have
// a locking mechanism to stay thread safe.
func (l *LRU) remove(listElement *list.Element) {
	l.linkedList.Remove(listElement)
	entryPtr := listElement.Value.(*entry)
	delete(l.kv, entryPtr.key)
	l.metrics.CacheRemoveInc()
}

// It is NOT thread safe and its parent should have
// a locking mechanism to stay thread safe.
func (l *LRU) removeOldest() {
	listElement := l.linkedList.Back()
	if listElement != nil {
		l.remove(listElement)
	}
}
