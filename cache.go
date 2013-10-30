package ccache

import (
  "time"
  "runtime"
  "hash/fnv"
  "container/list"
)

type Cache struct {
  *Configuration
  list *list.List
  buckets []*Bucket
  bucketCount uint32
  deletables chan *Item
  promotables chan *Item
}

func New(config *Configuration) *Cache {
  c := &Cache{
    list: new(list.List),
    Configuration: config,
    bucketCount: uint32(config.buckets),
    buckets: make([]*Bucket, config.buckets),
    deletables: make(chan *Item, config.deleteBuffer),
    promotables: make(chan *Item, config.promoteBuffer),
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
  bucket := c.bucket(key)
  item := bucket.get(key)
  if item == nil { return nil }
  if item.expires.Before(time.Now()) {
    c.deleteItem(bucket, item)
    return nil
  }
  c.promote(item)
  return item.value
}

func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
  item := c.bucket(key).set(key, value, duration)
  c.promote(item)
}

func (c *Cache) Fetch(key string, duration time.Duration, fetch func() interface{}) interface{} {
  item := c.Get(key)
  if item != nil { return item }
  value := fetch()
  c.Set(key, value, duration)
  return value
}

func (c *Cache) Delete(key string) {
  item := c.bucket(key).getAndDelete(key)
  if item != nil {
    c.deletables <- item
  }
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

func (c *Cache) promote(item *Item) {
  if item.shouldPromote(c.getsPerPromote) == false { return }
  c.promotables <- item
}

func (c *Cache) worker() {
  ms := new(runtime.MemStats)
  for {
    select {
    case item := <- c.promotables:
      wasNew := c.doPromote(item)
      if wasNew == false { continue }
      runtime.ReadMemStats(ms)
      if ms.HeapAlloc > c.size { c.gc() }
    case item := <- c.deletables:
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
  for i := 0; i < c.itemsToPrune; i++ {
    element := c.list.Back()
    if element == nil { return }
    item := element.Value.(*Item)
    c.bucket(item.key).delete(item.key)
    c.list.Remove(element)
  }
}
