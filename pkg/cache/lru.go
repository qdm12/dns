package cache

import (
	"container/list"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type lru struct {
	// Configuration
	maxEntries int
	ttlNano    int64

	// State
	kv         map[string]*list.Element
	linkedList *list.List
	mutex      sync.Mutex

	// Mock fields
	timeNow func() time.Time
}

func newLRU(maxEntries int, ttl time.Duration) *lru {
	return &lru{
		maxEntries: maxEntries,
		ttlNano:    int64(ttl),
		kv:         make(map[string]*list.Element, maxEntries),
		linkedList: list.New(),
		timeNow:    time.Now,
	}
}

func (l *lru) Add(request, response *dns.Msg) {
	if len(request.Question) == 0 {
		// cannot make key if there is no question
		return
	}

	key := makeKey(request)
	creationNano := l.timeNow().UnixNano()
	responseCopy := response.Copy()

	l.mutex.Lock()
	defer l.mutex.Unlock()

	if listElement, ok := l.kv[key]; ok {
		l.linkedList.MoveToFront(listElement)
		entryPtr := listElement.Value.(*entry)
		entryPtr.creationNano = creationNano
		entryPtr.response = responseCopy
		return
	}

	entry := &entry{
		key:          key,
		creationNano: creationNano,
		response:     responseCopy,
	}

	listElement := l.linkedList.PushFront(entry)
	l.kv[key] = listElement

	if l.maxEntries > 0 && l.linkedList.Len() > l.maxEntries {
		l.removeOldest()
	}
}

func (l *lru) Get(request *dns.Msg) (response *dns.Msg) {
	if len(request.Question) == 0 {
		// cannot make key if there is no question
		return
	}

	key := makeKey(request)
	nowNano := l.timeNow().UnixNano()

	l.mutex.Lock()
	defer l.mutex.Unlock()

	listElement, ok := l.kv[key]
	if !ok {
		return nil
	}

	l.linkedList.MoveToFront(listElement)
	entryPtr := listElement.Value.(*entry)

	if nowNano >= entryPtr.creationNano+l.ttlNano {
		// expired record
		l.remove(listElement)
		return nil
	}

	return entryPtr.response.Copy()
}

// remove removes a list element
// It is NOT thread safe and its parent should have
// a locking mechanism to stay thread safe.
func (l *lru) remove(listElement *list.Element) {
	l.linkedList.Remove(listElement)
	entryPtr := listElement.Value.(*entry)
	delete(l.kv, entryPtr.key)
}

// It is NOT thread safe and its parent should have
// a locking mechanism to stay thread safe.
func (l *lru) removeOldest() {
	listElement := l.linkedList.Back()
	if listElement != nil {
		l.remove(listElement)
	}
}
