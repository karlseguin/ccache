package ccache

import (
  "sync"
  "sync/atomic"
  "container/list"
)

const promoteCap = 5

type Item struct {
  key string
  value Value
  sync.RWMutex
  promotions int32
  element *list.Element
}

func newItem(key string, value Value) *Item {
  return &Item{
    key: key,
    value: value,
    promotions: -1,
  }
}

func (i *Item) shouldPromote() bool {
  promotions := atomic.AddInt32(&i.promotions, 1)
  if promotions == promoteCap || promotions == 0 {
    return true
  }
  return false
}
