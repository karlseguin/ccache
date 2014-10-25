package ccache

import (
	"container/list"
	"sync/atomic"
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
	promotions int32
	refCount   int32
	expires    int64
	value      interface{}
	element    *list.Element
}

func newItem(key string, value interface{}, expires int64) *Item {
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
