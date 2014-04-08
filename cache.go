// An LRU cached aimed at high concurrency
package ccache

import (
	"container/list"
	"hash/fnv"
	"sync/atomic"
	"time"
)

type Cache struct {
	*Configuration
	list        *list.List
	buckets     []*Bucket
	bucketCount uint32
	deletables  chan *Item
	promotables chan *Item
}

func New(config *Configuration) *Cache {
	c := &Cache{
		list:          list.New(),
		Configuration: config,
		bucketCount:   uint32(config.buckets),
		buckets:       make([]*Bucket, config.buckets),
		deletables:    make(chan *Item, config.deleteBuffer),
		promotables:   make(chan *Item, config.promoteBuffer),
	}
	for i := 0; i < config.buckets; i++ {
		c.buckets[i] = &Bucket{
			lookup: make(map[string]*Item),
		}
	}
	go c.worker()
	return c
}

func (c *Cache) Get(key string) interface{} {
	if item := c.get(key); item != nil {
		return item.value
	}
	return nil
}

func (c *Cache) TrackingGet(key string) TrackedItem {
	item := c.get(key)
	if item == nil {
		return NilTracked
	}
	item.track()
	return item
}

func (c *Cache) get(key string) *Item {
	bucket := c.bucket(key)
	item := bucket.get(key)
	if item == nil {
		return nil
	}
	if item.expires.Before(time.Now()) {
		c.deleteItem(bucket, item)
		return nil
	}
	c.conditionalPromote(item)
	return item
}

func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	item, new := c.bucket(key).set(key, value, duration)
	if new {
		c.promote(item)
	} else {
		c.conditionalPromote(item)
	}
}

func (c *Cache) Fetch(key string, duration time.Duration, fetch func() (interface{}, error)) (interface{}, error) {
	item := c.Get(key)
	if item != nil {
		return item, nil
	}
	value, err := fetch()
	if err == nil {
		c.Set(key, value, duration)
	}
	return value, err
}

func (c *Cache) Delete(key string) {
	item := c.bucket(key).getAndDelete(key)
	if item != nil {
		c.deletables <- item
	}
}

//this isn't thread safe. It's meant to be called from non-concurrent tests
func (c *Cache) Clear() {
	for _, bucket := range c.buckets {
		bucket.clear()
	}
	c.list = list.New()
}

func (c *Cache) deleteItem(bucket *Bucket, item *Item) {
	bucket.delete(item.key) //stop othe GETs from getting it
	c.deletables <- item
}

func (c *Cache) bucket(key string) *Bucket {
	h := fnv.New32a()
	h.Write([]byte(key))
	index := h.Sum32() % c.bucketCount
	return c.buckets[index]
}

func (c *Cache) conditionalPromote(item *Item) {
	if item.shouldPromote(c.getsPerPromote) == false {
		return
	}
	c.promote(item)
}

func (c *Cache) promote(item *Item) {
	c.promotables <- item
}

func (c *Cache) worker() {
	for {
		select {
		case item := <-c.promotables:
			if c.doPromote(item) && c.list.Len() > c.maxItems {
				c.gc()
			}
		case item := <-c.deletables:
			c.list.Remove(item.element)
		}
	}
}

func (c *Cache) doPromote(item *Item) bool {
	item.Lock()
	defer item.Unlock()
	item.promotions = 0
	if item.element != nil { //not a new item
		c.list.MoveToFront(item.element)
		return false
	}
	item.element = c.list.PushFront(item)
	return true
}

func (c *Cache) gc() {
	element := c.list.Back()
	for i := 0; i < c.itemsToPrune; i++ {
		if element == nil {
			return
		}
		prev := element.Prev()
		item := element.Value.(*Item)
		if c.tracking == false || atomic.LoadInt32(&item.refCount) == 0 {
			c.bucket(item.key).delete(item.key)
			c.list.Remove(element)
		}
		element = prev
	}
}
