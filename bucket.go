package ccache

import (
  "sync"
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

func (b *Bucket) set(key string, value Value) *Item {
  b.Lock()
  defer b.Unlock()
  if existing, exists := b.lookup[key]; exists {
    existing.Lock()
    existing.value = value
    existing.Unlock()
    return existing
  }
  item := newItem(key, value)
  b.lookup[key] = item
  return item
}


func (b *Bucket) remove(key string) {
  b.Lock()
  defer b.Unlock()
  delete(b.lookup, key)
}
