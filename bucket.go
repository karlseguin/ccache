package ccache

import (
	"sync"
	"time"
)

type Bucket struct {
	sync.RWMutex
	lookup map[string]*Item
}

func (b *Bucket) get(key string) *Item {
	b.RLock()
	defer b.RUnlock()
	return b.lookup[key]
}

func (b *Bucket) set(key string, value interface{}, duration time.Duration) (*Item, bool) {
	expires := time.Now().Add(duration)
	b.Lock()
	defer b.Unlock()
	if existing, exists := b.lookup[key]; exists {
		existing.Lock()
		existing.value = value
		existing.expires = expires
		existing.Unlock()
		return existing, false
	}
	item := newItem(key, value, expires)
	b.lookup[key] = item
	return item, true
}

func (b *Bucket) delete(key string) {
	b.Lock()
	defer b.Unlock()
	delete(b.lookup, key)
}

func (b *Bucket) getAndDelete(key string) *Item {
	b.Lock()
	defer b.Unlock()
	item := b.lookup[key]
	delete(b.lookup, key)
	return item
}

func (b *Bucket) clear() {
	b.Lock()
	defer b.Unlock()
	b.lookup = make(map[string]*Item)
}
