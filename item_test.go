package ccache

import (
  "testing"
  "github.com/karlseguin/gspec"
)

func TestItemPromotability(t *testing.T) {
  spec := gspec.New(t)
  item := &Item{promotions: -1}
  spec.Expect(item.shouldPromote()).ToEqual(true)
  spec.Expect(item.shouldPromote()).ToEqual(false)

  item.promotions = 4
  spec.Expect(item.shouldPromote()).ToEqual(true)
  spec.Expect(item.shouldPromote()).ToEqual(false)
}
