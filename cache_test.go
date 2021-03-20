package ccache

import (
	"sort"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/karlseguin/expect"
)

type CacheTests struct{}

func Test_Cache(t *testing.T) {
	Expectify(new(CacheTests), t)
}

func (_ CacheTests) DeletesAValue() {
	cache := New(Configure())
	defer cache.Stop()
	Expect(cache.ItemCount()).To.Equal(0)

	cache.Set("spice", "flow", time.Minute)
	cache.Set("worm", "sand", time.Minute)
	Expect(cache.ItemCount()).To.Equal(2)

	cache.Delete("spice")
	Expect(cache.Get("spice")).To.Equal(nil)
	Expect(cache.Get("worm").Value()).To.Equal("sand")
	Expect(cache.ItemCount()).To.Equal(1)
}

func (_ CacheTests) DeletesAPrefix() {
	cache := New(Configure())
	defer cache.Stop()
	Expect(cache.ItemCount()).To.Equal(0)

	cache.Set("aaa", "1", time.Minute)
	cache.Set("aab", "2", time.Minute)
	cache.Set("aac", "3", time.Minute)
	cache.Set("ac", "4", time.Minute)
	cache.Set("z5", "7", time.Minute)
	Expect(cache.ItemCount()).To.Equal(5)

	Expect(cache.DeletePrefix("9a")).To.Equal(0)
	Expect(cache.ItemCount()).To.Equal(5)

	Expect(cache.DeletePrefix("aa")).To.Equal(3)
	Expect(cache.Get("aaa")).To.Equal(nil)
	Expect(cache.Get("aab")).To.Equal(nil)
	Expect(cache.Get("aac")).To.Equal(nil)
	Expect(cache.Get("ac").Value()).To.Equal("4")
	Expect(cache.Get("z5").Value()).To.Equal("7")
	Expect(cache.ItemCount()).To.Equal(2)
}

func (_ CacheTests) DeletesAFunc() {
	cache := New(Configure())
	defer cache.Stop()
	Expect(cache.ItemCount()).To.Equal(0)

	cache.Set("a", 1, time.Minute)
	cache.Set("b", 2, time.Minute)
	cache.Set("c", 3, time.Minute)
	cache.Set("d", 4, time.Minute)
	cache.Set("e", 5, time.Minute)
	cache.Set("f", 6, time.Minute)
	Expect(cache.ItemCount()).To.Equal(6)

	Expect(cache.DeleteFunc(func(key string, item *Item) bool {
		return false
	})).To.Equal(0)
	Expect(cache.ItemCount()).To.Equal(6)

	Expect(cache.DeleteFunc(func(key string, item *Item) bool {
		return item.Value().(int) < 4
	})).To.Equal(3)
	Expect(cache.ItemCount()).To.Equal(3)

	Expect(cache.DeleteFunc(func(key string, item *Item) bool {
		return key == "d"
	})).To.Equal(1)
	Expect(cache.ItemCount()).To.Equal(2)

}

func (_ CacheTests) OnDeleteCallbackCalled() {
	onDeleteFnCalled := int32(0)
	onDeleteFn := func(item *Item) {
		if item.key == "spice" {
			atomic.AddInt32(&onDeleteFnCalled, 1)
		}
	}

	cache := New(Configure().OnDelete(onDeleteFn))
	cache.Set("spice", "flow", time.Minute)
	cache.Set("worm", "sand", time.Minute)

	cache.SyncUpdates() // wait for worker to pick up preceding updates

	cache.Delete("spice")
	cache.SyncUpdates()

	Expect(cache.Get("spice")).To.Equal(nil)
	Expect(cache.Get("worm").Value()).To.Equal("sand")
	Expect(atomic.LoadInt32(&onDeleteFnCalled)).To.Eql(1)
}

func (_ CacheTests) FetchesExpiredItems() {
	cache := New(Configure())
	fn := func() (interface{}, error) { return "moo-moo", nil }

	cache.Set("beef", "moo", time.Second*-1)
	Expect(cache.Get("beef").Value()).To.Equal("moo")

	out, _ := cache.Fetch("beef", time.Second, fn)
	Expect(out.Value()).To.Equal("moo-moo")
}

func (_ CacheTests) GCsTheOldestItems() {
	cache := New(Configure().ItemsToPrune(10))
	for i := 0; i < 500; i++ {
		cache.Set(strconv.Itoa(i), i, time.Minute)
	}
	cache.SyncUpdates()
	cache.GC()
	Expect(cache.Get("9")).To.Equal(nil)
	Expect(cache.Get("10").Value()).To.Equal(10)
	Expect(cache.ItemCount()).To.Equal(490)
}

func (_ CacheTests) PromotedItemsDontGetPruned() {
	cache := New(Configure().ItemsToPrune(10).GetsPerPromote(1))
	for i := 0; i < 500; i++ {
		cache.Set(strconv.Itoa(i), i, time.Minute)
	}
	cache.SyncUpdates()
	cache.Get("9")
	cache.SyncUpdates()
	cache.GC()
	Expect(cache.Get("9").Value()).To.Equal(9)
	Expect(cache.Get("10")).To.Equal(nil)
	Expect(cache.Get("11").Value()).To.Equal(11)
}

func (_ CacheTests) TrackerDoesNotCleanupHeldInstance() {
	cache := New(Configure().ItemsToPrune(11).Track())
	item0 := cache.TrackingSet("0", 0, time.Minute)
	for i := 1; i < 11; i++ {
		cache.Set(strconv.Itoa(i), i, time.Minute)
	}
	item1 := cache.TrackingGet("1")
	cache.SyncUpdates()
	cache.GC()
	Expect(cache.Get("0").Value()).To.Equal(0)
	Expect(cache.Get("1").Value()).To.Equal(1)
	item0.Release()
	item1.Release()
	cache.GC()
	Expect(cache.Get("0")).To.Equal(nil)
	Expect(cache.Get("1")).To.Equal(nil)
}

func (_ CacheTests) RemovesOldestItemWhenFull() {
	onDeleteFnCalled := false
	onDeleteFn := func(item *Item) {
		if item.key == "0" {
			onDeleteFnCalled = true
		}
	}

	cache := New(Configure().MaxSize(5).ItemsToPrune(1).OnDelete(onDeleteFn))
	for i := 0; i < 7; i++ {
		cache.Set(strconv.Itoa(i), i, time.Minute)
	}
	cache.SyncUpdates()
	Expect(cache.Get("0")).To.Equal(nil)
	Expect(cache.Get("1")).To.Equal(nil)
	Expect(cache.Get("2").Value()).To.Equal(2)
	Expect(onDeleteFnCalled).To.Equal(true)
	Expect(cache.ItemCount()).To.Equal(5)
}

func (_ CacheTests) RemovesOldestItemWhenFullBySizer() {
	cache := New(Configure().MaxSize(9).ItemsToPrune(2))
	for i := 0; i < 7; i++ {
		cache.Set(strconv.Itoa(i), &SizedItem{i, 2}, time.Minute)
	}
	cache.SyncUpdates()
	Expect(cache.Get("0")).To.Equal(nil)
	Expect(cache.Get("1")).To.Equal(nil)
	Expect(cache.Get("2")).To.Equal(nil)
	Expect(cache.Get("3")).To.Equal(nil)
	Expect(cache.Get("4").Value().(*SizedItem).id).To.Equal(4)
	Expect(cache.GetDropped()).To.Equal(4)
	Expect(cache.GetDropped()).To.Equal(0)
}

func (_ CacheTests) SetUpdatesSizeOnDelta() {
	cache := New(Configure())
	cache.Set("a", &SizedItem{0, 2}, time.Minute)
	cache.Set("b", &SizedItem{0, 3}, time.Minute)
	cache.SyncUpdates()
	Expect(cache.GetSize()).To.Eql(5)
	cache.Set("b", &SizedItem{0, 3}, time.Minute)
	cache.SyncUpdates()
	Expect(cache.GetSize()).To.Eql(5)
	cache.Set("b", &SizedItem{0, 4}, time.Minute)
	cache.SyncUpdates()
	Expect(cache.GetSize()).To.Eql(6)
	cache.Set("b", &SizedItem{0, 2}, time.Minute)
	cache.SyncUpdates()
	Expect(cache.GetSize()).To.Eql(4)
	cache.Delete("b")
	cache.SyncUpdates()
	Expect(cache.GetSize()).To.Eql(2)
}

func (_ CacheTests) ReplaceDoesNotchangeSizeIfNotSet() {
	cache := New(Configure())
	cache.Set("1", &SizedItem{1, 2}, time.Minute)
	cache.Set("2", &SizedItem{1, 2}, time.Minute)
	cache.Set("3", &SizedItem{1, 2}, time.Minute)
	cache.Replace("4", &SizedItem{1, 2})
	cache.SyncUpdates()
	Expect(cache.GetSize()).To.Eql(6)
}

func (_ CacheTests) ReplaceChangesSize() {
	cache := New(Configure())
	cache.Set("1", &SizedItem{1, 2}, time.Minute)
	cache.Set("2", &SizedItem{1, 2}, time.Minute)

	cache.Replace("2", &SizedItem{1, 2})
	cache.SyncUpdates()
	Expect(cache.GetSize()).To.Eql(4)

	cache.Replace("2", &SizedItem{1, 1})
	cache.SyncUpdates()
	Expect(cache.GetSize()).To.Eql(3)

	cache.Replace("2", &SizedItem{1, 3})
	cache.SyncUpdates()
	Expect(cache.GetSize()).To.Eql(5)
}

func (_ CacheTests) ResizeOnTheFly() {
	cache := New(Configure().MaxSize(9).ItemsToPrune(1))
	for i := 0; i < 5; i++ {
		cache.Set(strconv.Itoa(i), i, time.Minute)
	}
	cache.SetMaxSize(3)
	cache.SyncUpdates()
	Expect(cache.GetDropped()).To.Equal(2)
	Expect(cache.Get("0")).To.Equal(nil)
	Expect(cache.Get("1")).To.Equal(nil)
	Expect(cache.Get("2").Value()).To.Equal(2)
	Expect(cache.Get("3").Value()).To.Equal(3)
	Expect(cache.Get("4").Value()).To.Equal(4)

	cache.Set("5", 5, time.Minute)
	cache.SyncUpdates()
	Expect(cache.GetDropped()).To.Equal(1)
	Expect(cache.Get("2")).To.Equal(nil)
	Expect(cache.Get("3").Value()).To.Equal(3)
	Expect(cache.Get("4").Value()).To.Equal(4)
	Expect(cache.Get("5").Value()).To.Equal(5)

	cache.SetMaxSize(10)
	cache.Set("6", 6, time.Minute)
	cache.SyncUpdates()
	Expect(cache.GetDropped()).To.Equal(0)
	Expect(cache.Get("3").Value()).To.Equal(3)
	Expect(cache.Get("4").Value()).To.Equal(4)
	Expect(cache.Get("5").Value()).To.Equal(5)
	Expect(cache.Get("6").Value()).To.Equal(6)
}

func (_ CacheTests) ForEachFunc() {
	cache := New(Configure().MaxSize(3).ItemsToPrune(1))
	Expect(forEachKeys(cache)).To.Equal([]string{})

	cache.Set("1", 1, time.Minute)
	Expect(forEachKeys(cache)).To.Equal([]string{"1"})

	cache.Set("2", 2, time.Minute)
	cache.SyncUpdates()
	Expect(forEachKeys(cache)).To.Equal([]string{"1", "2"})

	cache.Set("3", 3, time.Minute)
	cache.SyncUpdates()
	Expect(forEachKeys(cache)).To.Equal([]string{"1", "2", "3"})

	cache.Set("4", 4, time.Minute)
	cache.SyncUpdates()
	Expect(forEachKeys(cache)).To.Equal([]string{"2", "3", "4"})

	cache.Set("stop", 5, time.Minute)
	cache.SyncUpdates()
	Expect(forEachKeys(cache)).Not.To.Contain("stop")

	cache.Set("6", 6, time.Minute)
	cache.SyncUpdates()
	Expect(forEachKeys(cache)).Not.To.Contain("stop")
}

type SizedItem struct {
	id int
	s  int64
}

func (s *SizedItem) Size() int64 {
	return s.s
}

func forEachKeys(cache *Cache) []string {
	keys := make([]string, 0, 10)
	cache.ForEachFunc(func(key string, i *Item) bool {
		if key == "stop" {
			return false
		}
		keys = append(keys, key)
		return true
	})
	sort.Strings(keys)
	return keys
}
