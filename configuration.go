package ccache

type Configuration struct {
	maxItems       int
	buckets        int
	itemsToPrune   int
	deleteBuffer   int
	promoteBuffer  int
	getsPerPromote int32
	tracking       bool
}

func Configure() *Configuration {
	return &Configuration{
		buckets:        16,
		itemsToPrune:   500,
		deleteBuffer:   1024,
		getsPerPromote: 3,
		promoteBuffer:  1024,
		maxItems:       5000,
		tracking:       false,
	}
}

// The size, in bytes, of the data to cache
// [5000]
func (c *Configuration) MaxItems(max int) *Configuration {
	c.maxItems = max
	return c
}

// Keys are hashed into % bucket count to provide greater concurrency (every set
// requires a write lock on the bucket)
// [64]
func (c *Configuration) Buckets(count int) *Configuration {
	c.buckets = count
	return c
}

// The number of items to prune when memory is low
// [500]
func (c *Configuration) ItemsToPrune(count int) *Configuration {
	c.itemsToPrune = count
	return c
}

// The size of the queue for items which should be promoted. If the queue fills
// up, promotions are skipped
// [1024]
func (c *Configuration) PromoteBuffer(size int) *Configuration {
	c.promoteBuffer = size
	return c
}

// The size of the queue for items which should be deleted. If the queue fills
// up, calls to Delete() will block
func (c *Configuration) DeleteBuffer(size int) *Configuration {
	c.deleteBuffer = size
	return c
}

// Give a large cache with a high read / write ratio, it's usually unecessary
// to promote an item on every Get. GetsPerPromote specifies the number of Gets
// a key must have before being promoted
// [10]
func (c *Configuration) GetsPerPromote(count int) *Configuration {
	c.getsPerPromote = int32(count)
	return c
}

// Typically, a cache is agnostic about how cached values are use. This is fine
// for a typical cache usage, where you fetch an item from the cache, do something
// (write it out to) and nothing else.

// However, if callers are going to keep a reference to a cached item for a long
// time, things get messy. Specifically, the cache can evict the item, while
// references still exist. Technically, this isn't an issue. However, if you reload
// the item back into the cache, you end up with 2 objects representing the same
// data. This is a waste of space and could lead to weird behavior (the type an
// identity map is meant to solve).

// By turning tracking on and using the cache's TrackingGet, the cache
// won't evict items which you haven't called Release() on. It's a simple reference
// counter.
func (c *Configuration) Track() *Configuration {
	c.tracking = true
	return c
}
