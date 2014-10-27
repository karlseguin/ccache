// An LRU cached aimed at high concurrency
package ccache

import (
	"container/list"
	"hash/fnv"
	"sync/atomic"
	"time"
)

type LayeredCache struct {
	*Configuration
	list        *list.List
	buckets     []*LayeredBucket
	bucketCount uint32
	deletables  chan *Item
	promotables chan *Item
}

func Layered(config *Configuration) *LayeredCache {
	c := &LayeredCache{
		list:          list.New(),
		Configuration: config,
		bucketCount:   uint32(config.buckets),
		buckets:       make([]*LayeredBucket, config.buckets),
		deletables:    make(chan *Item, config.deleteBuffer),
		promotables:   make(chan *Item, config.promoteBuffer),
	}
	for i := 0; i < int(config.buckets); i++ {
		c.buckets[i] = &LayeredBucket{
			buckets: make(map[string]*Bucket),
		}
	}
	go c.worker()
	return c
}

func (c *LayeredCache) Get(primary, secondary string) *Item {
	bucket := c.bucket(primary)
	item := bucket.get(primary, secondary)
	if item == nil {
		return nil
	}
	if item.expires > time.Now().Unix() {
		c.conditionalPromote(item)
	}
	return item
}

func (c *LayeredCache) TrackingGet(primary, secondary string) TrackedItem {
	item := c.Get(primary, secondary)
	if item == nil {
		return NilTracked
	}
	item.track()
	return item
}

func (c *LayeredCache) Set(primary, secondary string, value interface{}, duration time.Duration) {
	item, new := c.bucket(primary).set(primary, secondary, value, duration)
	if new {
		c.promote(item)
	} else {
		c.conditionalPromote(item)
	}
}

func (c *LayeredCache) Fetch(primary, secondary string, duration time.Duration, fetch func() (interface{}, error)) (interface{}, error) {
	item := c.Get(primary, secondary)
	if item != nil {
		return item, nil
	}
	value, err := fetch()
	if err == nil {
		c.Set(primary, secondary, value, duration)
	}
	return value, err
}

func (c *LayeredCache) Delete(primary, secondary string) bool {
	item := c.bucket(primary).delete(primary, secondary)
	if item != nil {
		c.deletables <- item
		return true
	}
	return false
}

func (c *LayeredCache) DeleteAll(primary string) bool {
	return c.bucket(primary).deleteAll(primary, c.deletables)
}

//this isn't thread safe. It's meant to be called from non-concurrent tests
func (c *LayeredCache) Clear() {
	for _, bucket := range c.buckets {
		bucket.clear()
	}
	c.list = list.New()
}

func (c *LayeredCache) bucket(key string) *LayeredBucket {
	h := fnv.New32a()
	h.Write([]byte(key))
	return c.buckets[h.Sum32()%c.bucketCount]
}

func (c *LayeredCache) conditionalPromote(item *Item) {
	if item.shouldPromote(c.getsPerPromote) == false {
		return
	}
	c.promote(item)
}

func (c *LayeredCache) promote(item *Item) {
	c.promotables <- item
}

func (c *LayeredCache) worker() {
	for {
		select {
		case item := <-c.promotables:
			if c.doPromote(item) && uint64(c.list.Len()) > c.maxItems {
				c.gc()
			}
		case item := <-c.deletables:
			if item.element == nil {
				item.promotions = -2
			} else {
				c.list.Remove(item.element)
			}
		}
	}
}

func (c *LayeredCache) doPromote(item *Item) bool {
	// deleted before it ever got promoted
	if item.promotions == -2 {
		return false
	}

	item.promotions = 0
	if item.element != nil { //not a new item
		c.list.MoveToFront(item.element)
		return false
	}
	item.element = c.list.PushFront(item)
	return true
}

func (c *LayeredCache) gc() {
	element := c.list.Back()
	for i := 0; i < c.itemsToPrune; i++ {
		if element == nil {
			return
		}
		prev := element.Prev()
		item := element.Value.(*Item)
		if c.tracking == false || atomic.LoadInt32(&item.refCount) == 0 {
			c.bucket(item.group).delete(item.group, item.key)
			c.list.Remove(element)
		}
		element = prev
	}
}
