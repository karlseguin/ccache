// An LRU cached aimed at high concurrency
package ccache

import (
	"hash/fnv"
	"sync/atomic"
	"time"
)

type Cache[T any] struct {
	*Configuration[T]
	control
	list            *List[T]
	size            int64
	pruneTargetSize int64
	buckets         []*bucket[T]
	bucketMask      uint32
	deletables      chan *Item[T]
	promotables     chan *Item[T]
}

// Create a new cache with the specified configuration
// See ccache.Configure() for creating a configuration
func New[T any](config *Configuration[T]) *Cache[T] {
	c := &Cache[T]{
		list:            NewList[T](),
		Configuration:   config,
		control:         newControl(),
		bucketMask:      uint32(config.buckets) - 1,
		buckets:         make([]*bucket[T], config.buckets),
		deletables:      make(chan *Item[T], config.deleteBuffer),
		promotables:     make(chan *Item[T], config.promoteBuffer),
		pruneTargetSize: config.maxSize - config.maxSize*int64(config.percentToPrune)/100,
	}
	for i := 0; i < config.buckets; i++ {
		c.buckets[i] = &bucket[T]{
			lookup: make(map[string]*Item[T]),
		}
	}
	go c.worker()
	return c
}

func (c *Cache[T]) ItemCount() int {
	count := 0
	for _, b := range c.buckets {
		count += b.itemCount()
	}
	return count
}

func (c *Cache[T]) DeletePrefix(prefix string) int {
	count := 0
	for _, b := range c.buckets {
		count += b.deletePrefix(prefix, c.deletables)
	}
	return count
}

// Deletes all items that the matches func evaluates to true.
func (c *Cache[T]) DeleteFunc(matches func(key string, item *Item[T]) bool) int {
	count := 0
	for _, b := range c.buckets {
		count += b.deleteFunc(matches, c.deletables)
	}
	return count
}

func (c *Cache[T]) ForEachFunc(matches func(key string, item *Item[T]) bool) {
	for _, b := range c.buckets {
		if !b.forEachFunc(matches) {
			break
		}
	}
}

// Get an item from the cache. Returns nil if the item wasn't found.
// This can return an expired item. Use item.Expired() to see if the item
// is expired and item.TTL() to see how long until the item expires (which
// will be negative for an already expired item).
func (c *Cache[T]) Get(key string) *Item[T] {
	item := c.bucket(key).get(key)
	if item == nil {
		return nil
	}
	if !item.Expired() {
		select {
		case c.promotables <- item:
		default:
		}
	}
	return item
}

// Same as Get but does not promote the value. This essentially circumvents the
// "least recently used" aspect of this cache. To some degree, it's akin to a
// "peak"
func (c *Cache[T]) GetWithoutPromote(key string) *Item[T] {
	return c.bucket(key).get(key)
}

// Used when the cache was created with the Track() configuration option.
// Avoid otherwise
func (c *Cache[T]) TrackingGet(key string) TrackedItem[T] {
	item := c.Get(key)
	if item == nil {
		return nil
	}
	item.track()
	return item
}

// Used when the cache was created with the Track() configuration option.
// Sets the item, and returns a tracked reference to it.
func (c *Cache[T]) TrackingSet(key string, value T, duration time.Duration) TrackedItem[T] {
	return c.set(key, value, duration, true)
}

// Set the value in the cache for the specified duration
func (c *Cache[T]) Set(key string, value T, duration time.Duration) {
	c.set(key, value, duration, false)
}

// Setnx set the value in the cache for the specified duration if not exists
func (c *Cache[T]) Setnx(key string, value T, duration time.Duration) {
	c.bucket(key).setnx(key, value, duration, false)
}

// Setnx2 set the value in the cache for the specified duration if not exists
func (c *Cache[T]) Setnx2(key string, f func() T, duration time.Duration) *Item[T] {
	item, existing := c.bucket(key).setnx2(key, f, duration, false)
	// consistent with Get
	if existing && !item.Expired() {
		select {
		case c.promotables <- item:
		default:
		}
		// consistent with set
	} else if !existing {
		c.promotables <- item
	}
	return item
}

// Replace the value if it exists, does not set if it doesn't.
// Returns true if the item existed an was replaced, false otherwise.
// Replace does not reset item's TTL
func (c *Cache[T]) Replace(key string, value T) bool {
	item := c.bucket(key).get(key)
	if item == nil {
		return false
	}
	c.Set(key, value, item.TTL())
	return true
}

// Extend the value if it exists, does not set if it doesn't exists.
// Returns true if the expire time of the item an was extended, false otherwise.
func (c *Cache[T]) Extend(key string, duration time.Duration) bool {
	item := c.bucket(key).get(key)
	if item == nil {
		return false
	}

	item.Extend(duration)
	return true
}

// Attempts to get the value from the cache and calles fetch on a miss (missing
// or stale item). If fetch returns an error, no value is cached and the error
// is returned back to the caller.
// Note that Fetch merely calls the public Get and Set functions. If you want
// a different Fetch behavior, such as thundering herd protection or returning
// expired items, implement it in your application.
func (c *Cache[T]) Fetch(key string, duration time.Duration, fetch func() (T, error)) (*Item[T], error) {
	item := c.Get(key)
	if item != nil && !item.Expired() {
		return item, nil
	}
	value, err := fetch()
	if err != nil {
		return nil, err
	}
	return c.set(key, value, duration, false), nil
}

// Remove the item from the cache, return true if the item was present, false otherwise.
func (c *Cache[T]) Delete(key string) bool {
	item := c.bucket(key).remove(key)
	if item != nil {
		c.deletables <- item
		return true
	}
	return false
}

func (c *Cache[T]) set(key string, value T, duration time.Duration, track bool) *Item[T] {
	item, existing := c.bucket(key).set(key, value, duration, track)
	if existing != nil {
		c.deletables <- existing
	}
	c.promotables <- item
	return item
}

func (c *Cache[T]) bucket(key string) *bucket[T] {
	h := fnv.New32a()
	h.Write([]byte(key))
	return c.buckets[h.Sum32()&c.bucketMask]
}

func (c *Cache[T]) halted(fn func()) {
	c.halt()
	defer c.unhalt()
	fn()
}

func (c *Cache[T]) halt() {
	for _, bucket := range c.buckets {
		bucket.Lock()
	}
}

func (c *Cache[T]) unhalt() {
	for _, bucket := range c.buckets {
		bucket.Unlock()
	}
}

func (c *Cache[T]) worker() {
	dropped := 0
	cc := c.control

	promoteItem := func(item *Item[T]) {
		if c.doPromote(item) && c.size > c.maxSize {
			dropped += c.gc()
		}
	}

	for {
		select {
		case item := <-c.promotables:
			promoteItem(item)
		case item := <-c.deletables:
			c.doDelete(item)
		case control := <-cc:
			switch msg := control.(type) {
			case controlStop:
				goto drain
			case controlGetDropped:
				msg.res <- dropped
				dropped = 0
			case controlSetMaxSize:
				newMaxSize := msg.size
				c.maxSize = newMaxSize
				c.pruneTargetSize = newMaxSize - newMaxSize*int64(c.percentToPrune)/100
				if c.size > c.maxSize {
					dropped += c.gc()
				}
				msg.done <- struct{}{}
			case controlClear:
				c.halted(func() {
					promotables := c.promotables
					for len(promotables) > 0 {
						<-promotables
					}
					deletables := c.deletables
					for len(deletables) > 0 {
						<-deletables
					}

					for _, bucket := range c.buckets {
						bucket.clear()
					}
					c.size = 0
					c.list = NewList[T]()
				})
				msg.done <- struct{}{}
			case controlGetSize:
				msg.res <- c.size
			case controlGC:
				dropped += c.gc()
				msg.done <- struct{}{}
			case controlSyncUpdates:
				doAllPendingPromotesAndDeletes(c.promotables, promoteItem, c.deletables, c.doDelete)
				msg.done <- struct{}{}
			}
		}
	}

drain:
	for {
		select {
		case item := <-c.deletables:
			c.doDelete(item)
		default:
			return
		}
	}
}

// This method is used to implement SyncUpdates. It simply receives and processes as many
// items as it can receive from the promotables and deletables channels immediately without
// blocking. If some other goroutine sends an item on either channel after this method has
// finished receiving, that's OK, because SyncUpdates only guarantees processing of values
// that were already sent by the same goroutine.
func doAllPendingPromotesAndDeletes[T any](
	promotables <-chan *Item[T],
	promoteFn func(*Item[T]),
	deletables <-chan *Item[T],
	deleteFn func(*Item[T]),
) {
doAllPromotes:
	for {
		select {
		case item := <-promotables:
			promoteFn(item)
		default:
			break doAllPromotes
		}
	}
doAllDeletes:
	for {
		select {
		case item := <-deletables:
			deleteFn(item)
		default:
			break doAllDeletes
		}
	}
}

func (c *Cache[T]) doDelete(item *Item[T]) {
	if item.next == nil && item.prev == nil {
		item.promotions = -2
	} else {
		c.size -= item.size
		if c.onDelete != nil {
			c.onDelete(item)
		}
		c.list.Remove(item)
		item.promotions = -2
	}
}

func (c *Cache[T]) doPromote(item *Item[T]) bool {
	//already deleted
	if item.promotions == -2 {
		return false
	}

	if item.next != nil || item.prev != nil { // not a new item
		if item.shouldPromote(c.getsPerPromote) {
			c.list.MoveToFront(item)
			item.promotions = 0
		}
		return false
	}

	c.size += item.size
	c.list.Insert(item)
	return true
}

func (c *Cache[T]) gc() int {
	dropped := 0
	item := c.list.Tail

	prunedSize := int64(0)
	sizeToPrune := c.size - c.pruneTargetSize

	for prunedSize < sizeToPrune {
		if item == nil {
			return dropped
		}
		// fmt.Println(item.key)
		prev := item.prev
		if !c.tracking || atomic.LoadInt32(&item.refCount) == 0 {
			c.bucket(item.key).delete(item.key)
			itemSize := item.size
			c.size -= itemSize
			prunedSize += itemSize

			c.list.Remove(item)
			if c.onDelete != nil {
				c.onDelete(item)
			}
			dropped += 1
			item.promotions = -2
		}
		item = prev
	}
	return dropped
}
