package ccache

import (
	. "github.com/karlseguin/expect"
	"testing"
)

type ItemTests struct{}

func Test_Item(t *testing.T) {
	Expectify(new(ItemTests), t)
}

func (i *ItemTests) Promotability() {
	item := &Item{promotions: 4}
	Expect(item.shouldPromote(5)).To.Equal(true)
	Expect(item.shouldPromote(5)).To.Equal(false)
}
