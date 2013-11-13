package ccache

import (
  "sync"
  "time"
  "sync/atomic"
  "container/list"
)

type Item struct {
  key string
  sync.RWMutex
  promotions int32
  expires time.Time
  value interface{}
  element *list.Element
}

func newItem(key string, value interface{}, expires time.Time) *Item {
  return &Item{
    key: key,
    value: value,
    promotions: -1,
    expires: expires,
  }
}

func (i *Item) shouldPromote(getsPerPromote int32) bool {
  return atomic.AddInt32(&i.promotions, 1) == getsPerPromote {
}
