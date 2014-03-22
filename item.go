package ccache

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

type TrackedItem interface {
	Value() interface{}
	Release()
}

type nilItem struct{}

func (n *nilItem) Value() interface{} { return nil }
func (n *nilItem) Release()           {}

var NilTracked = new(nilItem)

type Item struct {
	key string
	sync.RWMutex
	promotions int32
	refCount   int32
	expires    time.Time
	value      interface{}
	element    *list.Element
}

func newItem(key string, value interface{}, expires time.Time) *Item {
	return &Item{
		key:        key,
		value:      value,
		promotions: -1,
		expires:    expires,
	}
}

func (i *Item) shouldPromote(getsPerPromote int32) bool {
	return atomic.AddInt32(&i.promotions, 1) == getsPerPromote
}

func (i *Item) Value() interface{} {
	return i.value
}

func (i *Item) track() {
	atomic.AddInt32(&i.refCount, 1)
}

func (i *Item) Release() {
	atomic.AddInt32(&i.refCount, -1)
}
