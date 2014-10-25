package ccache

import (
	. "github.com/karlseguin/expect"
	"testing"
	"time"
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

func (i *ItemTests) Expired() {
	now := time.Now().Unix()
	item1 := &Item{expires: now + 1}
	item2 := &Item{expires: now - 1}
	Expect(item1.Expired()).To.Equal(false)
	Expect(item2.Expired()).To.Equal(true)
}

func (i *ItemTests) TTL() {
	now := time.Now().Unix()
	item1 := &Item{expires: now + 10}
	item2 := &Item{expires: now - 10}
	Expect(item1.TTL()).To.Equal(time.Second * 10)
	Expect(item2.TTL()).To.Equal(time.Second * -10)
}


func (i *ItemTests) Expires() {
	now := time.Now().Unix()
	item1 := &Item{expires: now + 10}
	Expect(item1.Expires().Unix()).To.Equal(now + 10)
}
