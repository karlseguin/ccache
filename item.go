package ccache

import (
  "sync"
  "time"
  "container/list"
)

type Item struct {
  key string
  value Value
  sync.RWMutex
  promoted time.Time
  element *list.Element
}

func (i *Item) shouldPromote(staleness time.Duration) bool {
  stale := time.Now().Add(staleness)
  i.RLock()
  defer i.RUnlock()
  return i.promoted.Before(stale)
}
