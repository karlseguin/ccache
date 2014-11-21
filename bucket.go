package ccache

import (
	"sync"
	"sync/atomic"
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

func (b *bucket) set(key string, value interface{}, duration time.Duration) (*Item, bool, int64) {
	expires := time.Now().Add(duration).Unix()
	b.Lock()
	defer b.Unlock()
	if existing, exists := b.lookup[key]; exists {
		s := existing.size
		existing.value = value
		existing.expires = expires
		d := int64(0)
		if sized, ok := value.(Sized); ok {
			newSize := sized.Size()
			d = newSize - s
			if d != 0 {
				atomic.StoreInt64(&existing.size, newSize)
			}
		}
		return existing, false, int64(d)
	}
	item := newItem(key, value, expires)
	b.lookup[key] = item
	return item, true, int64(item.size)
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
