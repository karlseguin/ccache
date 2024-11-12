package ccache

import (
	"testing"

	"github.com/karlseguin/ccache/v3/assert"
)

func Test_List_Insert(t *testing.T) {
	l := NewList[int]()
	assertList(t, l)

	l.Insert(newItem("a", 1, 0, false))
	assertList(t, l, 1)

	l.Insert(newItem("b", 2, 0, false))
	assertList(t, l, 2, 1)

	l.Insert(newItem("c", 3, 0, false))
	assertList(t, l, 3, 2, 1)
}

func Test_List_Remove(t *testing.T) {
	l := NewList[int]()
	assertList(t, l)

	item := newItem("a", 1, 0, false)
	l.Insert(item)
	l.Remove(item)
	assertList(t, l)

	n5 := newItem("e", 5, 0, false)
	l.Insert(n5)
	n4 := newItem("d", 4, 0, false)
	l.Insert(n4)
	n3 := newItem("c", 3, 0, false)
	l.Insert(n3)
	n2 := newItem("b", 2, 0, false)
	l.Insert(n2)
	n1 := newItem("a", 1, 0, false)
	l.Insert(n1)

	l.Remove(n5)
	assertList(t, l, 1, 2, 3, 4)

	l.Remove(n1)
	assertList(t, l, 2, 3, 4)

	l.Remove(n3)
	assertList(t, l, 2, 4)

	l.Remove(n2)
	assertList(t, l, 4)

	l.Remove(n4)
	assertList(t, l)
}

func assertList(t *testing.T, list *List[int], expected ...int) {
	t.Helper()

	if len(expected) == 0 {
		assert.Nil(t, list.Head)
		assert.Nil(t, list.Tail)
		return
	}

	node := list.Head
	for _, expected := range expected {
		assert.Equal(t, node.value, expected)
		node = node.next
	}

	node = list.Tail
	for i := len(expected) - 1; i >= 0; i-- {
		assert.Equal(t, node.value, expected[i])
		node = node.prev
	}
}
