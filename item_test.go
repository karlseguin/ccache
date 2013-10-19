package ccache

import (
  "time"
  "testing"
  "github.com/karlseguin/gspec"
)

func TestItemPromotability(t *testing.T) {
  spec := gspec.New(t)
  item := &Item{promoted: time.Now().Add(time.Second * -5)}
  spec.Expect(item.shouldPromote(time.Second * -2)).ToEqual(true)
  spec.Expect(item.shouldPromote(time.Second * -6)).ToEqual(false)
}
