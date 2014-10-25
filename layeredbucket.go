package ccache

import (
	"sync"
	"time"
)

type LayeredBucket struct {
	sync.RWMutex
	buckets map[string]*Bucket
}

func (b *LayeredBucket) get(primary, secondary string) *Item {
	b.RLock()
	bucket, exists := b.buckets[primary]
	b.RUnlock()
	if exists == false {
		return nil
	}
	return bucket.get(secondary)
}

func (b *LayeredBucket) set(primary, secondary string, value interface{}, duration time.Duration) (*Item, bool) {
	b.Lock()
	bucket, exists := b.buckets[primary]
	if exists == false {
		bucket = &Bucket{lookup: make(map[string]*Item)}
		b.buckets[primary] = bucket
	}
	b.Unlock()
	item, new := bucket.set(secondary, value, duration)
	if new {
		item.group = primary
	}
	return item, new
}

func (b *LayeredBucket) delete(primary, secondary string) *Item {
	b.RLock()
	bucket, exists := b.buckets[primary]
	b.RUnlock()
	if exists == false {
		return nil
	}
	return bucket.delete(secondary)
}

func (b *LayeredBucket) deleteAll(primary string, deletables chan *Item) {
	b.RLock()
	bucket, exists := b.buckets[primary]
	b.RUnlock()
	if exists == false {
		return
	}

	bucket.Lock()
	defer bucket.Unlock()

	if l := len(bucket.lookup); l == 0 {
		return
	}
	for key, item := range bucket.lookup {
		delete(bucket.lookup, key)
		deletables <- item
	}
}

func (b *LayeredBucket) clear() {
	b.Lock()
	defer b.Unlock()
	for _, bucket := range b.buckets {
		bucket.clear()
	}
	b.buckets = make(map[string]*Bucket)
}
