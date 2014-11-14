package ccache

import (
	"sync"
	"time"
)

type bucket struct {
	sync.RWMutex
	lookup map[string]*Item
}

func (b *bucket) get(key string) *Item {
	b.RLock()
	defer b.RUnlock()
	return b.lookup[key]
}

func (b *bucket) set(key string, value interface{}, duration time.Duration) (*Item, bool) {
	expires := time.Now().Add(duration).Unix()
	b.Lock()
	defer b.Unlock()
	if existing, exists := b.lookup[key]; exists {
		existing.value = value
		existing.expires = expires
		return existing, false
	}
	item := newItem(key, value, expires)
	b.lookup[key] = item
	return item, true
}

func (b *bucket) replace(key string, value interface{}) bool {
	b.Lock()
	defer b.Unlock()
	existing, exists := b.lookup[key]
	if exists == false {
		return false
	}
	existing.value = value
	return true
}

func (b *bucket) delete(key string) *Item {
	b.Lock()
	defer b.Unlock()
	item := b.lookup[key]
	delete(b.lookup, key)
	return item
}

func (b *bucket) clear() {
	b.Lock()
	defer b.Unlock()
	b.lookup = make(map[string]*Item)
}
