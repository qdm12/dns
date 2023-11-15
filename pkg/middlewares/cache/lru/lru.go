package lru

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type LRU struct {
	// Configuration
	maxEntries uint

	// State
	kv         map[string]*list.Element
	linkedList *list.List
	mutex      sync.Mutex

	// External objects
	metrics Metrics

	// Mock fields
	timeNow func() time.Time
}

func New(settings Settings) (cache *LRU, err error) {
	settings.SetDefaults()

	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("settings validation: %w", err)
	}

	settings.Metrics.SetCacheType(CacheType)
	settings.Metrics.CacheMaxEntriesSet(settings.MaxEntries)

	return &LRU{
		maxEntries: settings.MaxEntries,
		kv:         make(map[string]*list.Element, settings.MaxEntries),
		linkedList: list.New(),
		metrics:    settings.Metrics,
		timeNow:    time.Now,
	}, nil
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
		entryPtr := listElement.Value.(*entry) //nolint:forcetypeassert
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

	if l.maxEntries > 0 && l.linkedList.Len() == int(l.maxEntries) {
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

	l.mutex.Lock()
	defer l.mutex.Unlock()

	listElement, ok := l.kv[key]
	if !ok {
		l.metrics.CacheMissInc()
		return nil
	}

	l.metrics.CacheHitInc()

	l.linkedList.MoveToFront(listElement)
	entryPtr := listElement.Value.(*entry) //nolint:forcetypeassert

	if l.timeNow().Unix() >= entryPtr.expUnix {
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
	entryPtr := listElement.Value.(*entry) //nolint:forcetypeassert
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
