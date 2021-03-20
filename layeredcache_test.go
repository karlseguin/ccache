package ccache

import (
	"sort"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/karlseguin/expect"
)

type LayeredCacheTests struct{}

func Test_LayeredCache(t *testing.T) {
	Expectify(new(LayeredCacheTests), t)
}

func (_ *LayeredCacheTests) GetsANonExistantValue() {
	cache := newLayered()
	Expect(cache.Get("spice", "flow")).To.Equal(nil)
	Expect(cache.ItemCount()).To.Equal(0)
}

func (_ *LayeredCacheTests) SetANewValue() {
	cache := newLayered()
	cache.Set("spice", "flow", "a value", time.Minute)
	Expect(cache.Get("spice", "flow").Value()).To.Equal("a value")
	Expect(cache.Get("spice", "stop")).To.Equal(nil)
	Expect(cache.ItemCount()).To.Equal(1)
}

func (_ *LayeredCacheTests) SetsMultipleValueWithinTheSameLayer() {
	cache := newLayered()
	cache.Set("spice", "flow", "value-a", time.Minute)
	cache.Set("spice", "must", "value-b", time.Minute)
	cache.Set("leto", "sister", "ghanima", time.Minute)
	Expect(cache.Get("spice", "flow").Value()).To.Equal("value-a")
	Expect(cache.Get("spice", "must").Value()).To.Equal("value-b")
	Expect(cache.Get("spice", "worm")).To.Equal(nil)

	Expect(cache.Get("leto", "sister").Value()).To.Equal("ghanima")
	Expect(cache.Get("leto", "brother")).To.Equal(nil)
	Expect(cache.Get("baron", "friend")).To.Equal(nil)
	Expect(cache.ItemCount()).To.Equal(3)
}

func (_ *LayeredCacheTests) ReplaceDoesNothingIfKeyDoesNotExist() {
	cache := newLayered()
	Expect(cache.Replace("spice", "flow", "value-a")).To.Equal(false)
	Expect(cache.Get("spice", "flow")).To.Equal(nil)
}

func (_ *LayeredCacheTests) ReplaceUpdatesTheValue() {
	cache := newLayered()
	cache.Set("spice", "flow", "value-a", time.Minute)
	Expect(cache.Replace("spice", "flow", "value-b")).To.Equal(true)
	Expect(cache.Get("spice", "flow").Value().(string)).To.Equal("value-b")
	Expect(cache.ItemCount()).To.Equal(1)
	//not sure how to test that the TTL hasn't changed sort of a sleep..
}

func (_ *LayeredCacheTests) DeletesAValue() {
	cache := newLayered()
	cache.Set("spice", "flow", "value-a", time.Minute)
	cache.Set("spice", "must", "value-b", time.Minute)
	cache.Set("leto", "sister", "ghanima", time.Minute)
	cache.Delete("spice", "flow")
	Expect(cache.Get("spice", "flow")).To.Equal(nil)
	Expect(cache.Get("spice", "must").Value()).To.Equal("value-b")
	Expect(cache.Get("spice", "worm")).To.Equal(nil)
	Expect(cache.Get("leto", "sister").Value()).To.Equal("ghanima")
	Expect(cache.ItemCount()).To.Equal(2)
}

func (_ *LayeredCacheTests) DeletesAPrefix() {
	cache := newLayered()
	Expect(cache.ItemCount()).To.Equal(0)

	cache.Set("spice", "aaa", "1", time.Minute)
	cache.Set("spice", "aab", "2", time.Minute)
	cache.Set("spice", "aac", "3", time.Minute)
	cache.Set("leto", "aac", "3", time.Minute)
	cache.Set("spice", "ac", "4", time.Minute)
	cache.Set("spice", "z5", "7", time.Minute)
	Expect(cache.ItemCount()).To.Equal(6)

	Expect(cache.DeletePrefix("spice", "9a")).To.Equal(0)
	Expect(cache.ItemCount()).To.Equal(6)

	Expect(cache.DeletePrefix("spice", "aa")).To.Equal(3)
	Expect(cache.Get("spice", "aaa")).To.Equal(nil)
	Expect(cache.Get("spice", "aab")).To.Equal(nil)
	Expect(cache.Get("spice", "aac")).To.Equal(nil)
	Expect(cache.Get("spice", "ac").Value()).To.Equal("4")
	Expect(cache.Get("spice", "z5").Value()).To.Equal("7")
	Expect(cache.ItemCount()).To.Equal(3)
}

func (_ *LayeredCacheTests) DeletesAFunc() {
	cache := newLayered()
	Expect(cache.ItemCount()).To.Equal(0)

	cache.Set("spice", "a", 1, time.Minute)
	cache.Set("leto", "b", 2, time.Minute)
	cache.Set("spice", "c", 3, time.Minute)
	cache.Set("spice", "d", 4, time.Minute)
	cache.Set("spice", "e", 5, time.Minute)
	cache.Set("spice", "f", 6, time.Minute)
	Expect(cache.ItemCount()).To.Equal(6)

	Expect(cache.DeleteFunc("spice", func(key string, item *Item) bool {
		return false
	})).To.Equal(0)
	Expect(cache.ItemCount()).To.Equal(6)

	Expect(cache.DeleteFunc("spice", func(key string, item *Item) bool {
		return item.Value().(int) < 4
	})).To.Equal(2)
	Expect(cache.ItemCount()).To.Equal(4)

	Expect(cache.DeleteFunc("spice", func(key string, item *Item) bool {
		return key == "d"
	})).To.Equal(1)
	Expect(cache.ItemCount()).To.Equal(3)

}

func (_ *LayeredCacheTests) OnDeleteCallbackCalled() {
	onDeleteFnCalled := int32(0)
	onDeleteFn := func(item *Item) {
		if item.group == "spice" && item.key == "flow" {
			atomic.AddInt32(&onDeleteFnCalled, 1)
		}
	}

	cache := Layered(Configure().OnDelete(onDeleteFn))
	cache.Set("spice", "flow", "value-a", time.Minute)
	cache.Set("spice", "must", "value-b", time.Minute)
	cache.Set("leto", "sister", "ghanima", time.Minute)

	cache.SyncUpdates()
	cache.Delete("spice", "flow")
	cache.SyncUpdates()

	Expect(cache.Get("spice", "flow")).To.Equal(nil)
	Expect(cache.Get("spice", "must").Value()).To.Equal("value-b")
	Expect(cache.Get("spice", "worm")).To.Equal(nil)
	Expect(cache.Get("leto", "sister").Value()).To.Equal("ghanima")

	Expect(atomic.LoadInt32(&onDeleteFnCalled)).To.Eql(1)
}

func (_ *LayeredCacheTests) DeletesALayer() {
	cache := newLayered()
	cache.Set("spice", "flow", "value-a", time.Minute)
	cache.Set("spice", "must", "value-b", time.Minute)
	cache.Set("leto", "sister", "ghanima", time.Minute)
	cache.DeleteAll("spice")
	Expect(cache.Get("spice", "flow")).To.Equal(nil)
	Expect(cache.Get("spice", "must")).To.Equal(nil)
	Expect(cache.Get("spice", "worm")).To.Equal(nil)
	Expect(cache.Get("leto", "sister").Value()).To.Equal("ghanima")
}

func (_ LayeredCacheTests) GCsTheOldestItems() {
	cache := Layered(Configure().ItemsToPrune(10))
	cache.Set("xx", "a", 23, time.Minute)
	for i := 0; i < 500; i++ {
		cache.Set(strconv.Itoa(i), "a", i, time.Minute)
	}
	cache.Set("xx", "b", 9001, time.Minute)
	//let the items get promoted (and added to our list)
	cache.SyncUpdates()
	cache.GC()
	Expect(cache.Get("xx", "a")).To.Equal(nil)
	Expect(cache.Get("xx", "b").Value()).To.Equal(9001)
	Expect(cache.Get("8", "a")).To.Equal(nil)
	Expect(cache.Get("9", "a").Value()).To.Equal(9)
	Expect(cache.Get("10", "a").Value()).To.Equal(10)
}

func (_ LayeredCacheTests) PromotedItemsDontGetPruned() {
	cache := Layered(Configure().ItemsToPrune(10).GetsPerPromote(1))
	for i := 0; i < 500; i++ {
		cache.Set(strconv.Itoa(i), "a", i, time.Minute)
	}
	cache.SyncUpdates()
	cache.Get("9", "a")
	cache.SyncUpdates()
	cache.GC()
	Expect(cache.Get("9", "a").Value()).To.Equal(9)
	Expect(cache.Get("10", "a")).To.Equal(nil)
	Expect(cache.Get("11", "a").Value()).To.Equal(11)
}

func (_ LayeredCacheTests) TrackerDoesNotCleanupHeldInstance() {
	cache := Layered(Configure().ItemsToPrune(10).Track())
	item0 := cache.TrackingSet("0", "a", 0, time.Minute)
	for i := 1; i < 11; i++ {
		cache.Set(strconv.Itoa(i), "a", i, time.Minute)
	}
	item1 := cache.TrackingGet("1", "a")
	cache.SyncUpdates()
	cache.GC()
	Expect(cache.Get("0", "a").Value()).To.Equal(0)
	Expect(cache.Get("1", "a").Value()).To.Equal(1)
	item0.Release()
	item1.Release()
	cache.GC()
	Expect(cache.Get("0", "a")).To.Equal(nil)
	Expect(cache.Get("1", "a")).To.Equal(nil)
}

func (_ LayeredCacheTests) RemovesOldestItemWhenFull() {
	cache := Layered(Configure().MaxSize(5).ItemsToPrune(1))
	cache.Set("xx", "a", 23, time.Minute)
	for i := 0; i < 7; i++ {
		cache.Set(strconv.Itoa(i), "a", i, time.Minute)
	}
	cache.Set("xx", "b", 9001, time.Minute)
	cache.SyncUpdates()
	Expect(cache.Get("xx", "a")).To.Equal(nil)
	Expect(cache.Get("0", "a")).To.Equal(nil)
	Expect(cache.Get("1", "a")).To.Equal(nil)
	Expect(cache.Get("2", "a")).To.Equal(nil)
	Expect(cache.Get("3", "a").Value()).To.Equal(3)
	Expect(cache.Get("xx", "b").Value()).To.Equal(9001)
	Expect(cache.GetDropped()).To.Equal(4)
	Expect(cache.GetDropped()).To.Equal(0)
}

func (_ LayeredCacheTests) ResizeOnTheFly() {
	cache := Layered(Configure().MaxSize(9).ItemsToPrune(1))
	for i := 0; i < 5; i++ {
		cache.Set(strconv.Itoa(i), "a", i, time.Minute)
	}
	cache.SyncUpdates()

	cache.SetMaxSize(3)
	cache.SyncUpdates()
	Expect(cache.GetDropped()).To.Equal(2)
	Expect(cache.Get("0", "a")).To.Equal(nil)
	Expect(cache.Get("1", "a")).To.Equal(nil)
	Expect(cache.Get("2", "a").Value()).To.Equal(2)
	Expect(cache.Get("3", "a").Value()).To.Equal(3)
	Expect(cache.Get("4", "a").Value()).To.Equal(4)

	cache.Set("5", "a", 5, time.Minute)
	cache.SyncUpdates()
	Expect(cache.GetDropped()).To.Equal(1)
	Expect(cache.Get("2", "a")).To.Equal(nil)
	Expect(cache.Get("3", "a").Value()).To.Equal(3)
	Expect(cache.Get("4", "a").Value()).To.Equal(4)
	Expect(cache.Get("5", "a").Value()).To.Equal(5)

	cache.SetMaxSize(10)
	cache.Set("6", "a", 6, time.Minute)
	cache.SyncUpdates()
	Expect(cache.GetDropped()).To.Equal(0)
	Expect(cache.Get("3", "a").Value()).To.Equal(3)
	Expect(cache.Get("4", "a").Value()).To.Equal(4)
	Expect(cache.Get("5", "a").Value()).To.Equal(5)
	Expect(cache.Get("6", "a").Value()).To.Equal(6)
}

func (_ LayeredCacheTests) RemovesOldestItemWhenFullBySizer() {
	cache := Layered(Configure().MaxSize(9).ItemsToPrune(2))
	for i := 0; i < 7; i++ {
		cache.Set("pri", strconv.Itoa(i), &SizedItem{i, 2}, time.Minute)
	}
	cache.SyncUpdates()
	Expect(cache.Get("pri", "0")).To.Equal(nil)
	Expect(cache.Get("pri", "1")).To.Equal(nil)
	Expect(cache.Get("pri", "2")).To.Equal(nil)
	Expect(cache.Get("pri", "3")).To.Equal(nil)
	Expect(cache.Get("pri", "4").Value().(*SizedItem).id).To.Equal(4)
}

func (_ LayeredCacheTests) SetUpdatesSizeOnDelta() {
	cache := Layered(Configure())
	cache.Set("pri", "a", &SizedItem{0, 2}, time.Minute)
	cache.Set("pri", "b", &SizedItem{0, 3}, time.Minute)
	cache.SyncUpdates()
	Expect(cache.GetSize()).To.Eql(5)
	cache.Set("pri", "b", &SizedItem{0, 3}, time.Minute)
	cache.SyncUpdates()
	Expect(cache.GetSize()).To.Eql(5)
	cache.Set("pri", "b", &SizedItem{0, 4}, time.Minute)
	cache.SyncUpdates()
	Expect(cache.GetSize()).To.Eql(6)
	cache.Set("pri", "b", &SizedItem{0, 2}, time.Minute)
	cache.Set("sec", "b", &SizedItem{0, 3}, time.Minute)
	cache.SyncUpdates()
	Expect(cache.GetSize()).To.Eql(7)
	cache.Delete("pri", "b")
	cache.SyncUpdates()
	Expect(cache.GetSize()).To.Eql(5)
}

func (_ LayeredCacheTests) ReplaceDoesNotchangeSizeIfNotSet() {
	cache := Layered(Configure())
	cache.Set("pri", "1", &SizedItem{1, 2}, time.Minute)
	cache.Set("pri", "2", &SizedItem{1, 2}, time.Minute)
	cache.Set("pri", "3", &SizedItem{1, 2}, time.Minute)
	cache.Replace("sec", "3", &SizedItem{1, 2})
	cache.SyncUpdates()
	Expect(cache.GetSize()).To.Eql(6)
}

func (_ LayeredCacheTests) ReplaceChangesSize() {
	cache := Layered(Configure())
	cache.Set("pri", "1", &SizedItem{1, 2}, time.Minute)
	cache.Set("pri", "2", &SizedItem{1, 2}, time.Minute)

	cache.Replace("pri", "2", &SizedItem{1, 2})
	cache.SyncUpdates()
	Expect(cache.GetSize()).To.Eql(4)

	cache.Replace("pri", "2", &SizedItem{1, 1})
	cache.SyncUpdates()
	Expect(cache.GetSize()).To.Eql(3)

	cache.Replace("pri", "2", &SizedItem{1, 3})
	cache.SyncUpdates()
	Expect(cache.GetSize()).To.Eql(5)
}

func (_ LayeredCacheTests) EachFunc() {
	cache := Layered(Configure().MaxSize(3).ItemsToPrune(1))
	Expect(forEachKeysLayered(cache, "1")).To.Equal([]string{})

	cache.Set("1", "a", 1, time.Minute)
	Expect(forEachKeysLayered(cache, "1")).To.Equal([]string{"a"})

	cache.Set("1", "b", 2, time.Minute)
	cache.SyncUpdates()
	Expect(forEachKeysLayered(cache, "1")).To.Equal([]string{"a", "b"})

	cache.Set("1", "c", 3, time.Minute)
	cache.SyncUpdates()
	Expect(forEachKeysLayered(cache, "1")).To.Equal([]string{"a", "b", "c"})

	cache.Set("1", "d", 4, time.Minute)
	cache.SyncUpdates()
	Expect(forEachKeysLayered(cache, "1")).To.Equal([]string{"b", "c", "d"})

	// iteration is non-deterministic, all we know for sure is "stop" should not be in there
	cache.Set("1", "stop", 5, time.Minute)
	cache.SyncUpdates()
	Expect(forEachKeysLayered(cache, "1")).Not.To.Contain("stop")

	cache.Set("1", "e", 6, time.Minute)
	cache.SyncUpdates()
	Expect(forEachKeysLayered(cache, "1")).Not.To.Contain("stop")
}

func newLayered() *LayeredCache {
	c := Layered(Configure())
	c.Clear()
	return c
}

func forEachKeysLayered(cache *LayeredCache, primary string) []string {
	keys := make([]string, 0, 10)
	cache.ForEachFunc(primary, func(key string, i *Item) bool {
		if key == "stop" {
			return false
		}
		keys = append(keys, key)
		return true
	})
	sort.Strings(keys)
	return keys
}
