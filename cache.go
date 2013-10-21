package ccache

import (
  "time"
  "runtime"
  "hash/fnv"
  "container/list"
)

type Value interface {
  Expires() time.Time
}

type Cache struct {
  *Configuration
  list *list.List
  buckets []*Bucket
  bucketCount uint32
  promotables chan *Item
}

func New(config *Configuration) *Cache {
  c := &Cache{
    list: new(list.List),
    Configuration: config,
    bucketCount: uint32(config.buckets),
    buckets: make([]*Bucket, config.buckets),
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

func (c *Cache) Get(key string) Value {
  item := c.bucket(key).get(key)
  if item == nil { return nil }
  c.promote(item)
  return item.value
}

func (c *Cache) Set(key string, value Value) {
  item := c.bucket(key).set(key, value)
  c.promote(item)
}

func (c *Cache) bucket(key string) *Bucket {
  h := fnv.New32a()
  h.Write([]byte(key))
  index := h.Sum32() % c.bucketCount
  return c.buckets[index]
}

func (c *Cache) promote(item *Item) {
  if item.shouldPromote() == false { return }
  c.promotables <- item
}

func (c *Cache) worker() {
  ms := new(runtime.MemStats)
  for {
    wasNew := c.doPromote(<- c.promotables)
    if wasNew == false { continue }
    runtime.ReadMemStats(ms)
    if ms.HeapAlloc > c.size{
      c.gc()
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
    c.bucket(item.key).remove(item.key)
    c.list.Remove(element)
  }
}
