package ccache

import (
	"github.com/karlseguin/gspec"
	"strconv"
	"testing"
	"time"
)

func TestCacheGCsTheOldestItems(t *testing.T) {
	spec := gspec.New(t)
	cache := New(Configure().ItemsToPrune(10))
	for i := 0; i < 500; i++ {
		cache.Set(strconv.Itoa(i), i, time.Minute)
	}
	cache.gc()
	spec.Expect(cache.Get("9")).ToBeNil()
	spec.Expect(cache.Get("10").(int)).ToEqual(10)
}

func TestCachePromotedItemsDontGetPruned(t *testing.T) {
	spec := gspec.New(t)
	cache := New(Configure().ItemsToPrune(10).GetsPerPromote(1))
	for i := 0; i < 500; i++ {
		cache.Set(strconv.Itoa(i), i, time.Minute)
	}
	time.Sleep(time.Millisecond * 10) //run the worker once to init the list
	cache.Get("9")
	time.Sleep(time.Millisecond * 10)
	cache.gc()
	spec.Expect(cache.Get("9").(int)).ToEqual(9)
	spec.Expect(cache.Get("10")).ToBeNil()
	spec.Expect(cache.Get("11").(int)).ToEqual(11)
}

func TestCacheTrackerDoesNotCleanupHeldInstance(t *testing.T) {
	spec := gspec.New(t)
	cache := New(Configure().ItemsToPrune(10).Track())
	for i := 0; i < 10; i++ {
		cache.Set(strconv.Itoa(i), i, time.Minute)
	}
	item := cache.TrackingGet("0")
	time.Sleep(time.Millisecond * 10)
	cache.gc()
	spec.Expect(cache.Get("0").(int)).ToEqual(0)
	spec.Expect(cache.Get("1")).ToBeNil()
	item.Release()
	cache.gc()
	spec.Expect(cache.Get("0")).ToBeNil()
}

func TestCacheRemovesOldestItemWhenFull(t *testing.T) {
	spec := gspec.New(t)
	cache := New(Configure().MaxItems(5).ItemsToPrune(1))
	for i := 0; i < 7; i++ {
		cache.Set(strconv.Itoa(i), i, time.Minute)
	}
	time.Sleep(time.Millisecond * 10)
	spec.Expect(cache.Get("0")).ToBeNil()
	spec.Expect(cache.Get("1")).ToBeNil()
	spec.Expect(cache.Get("2").(int)).ToEqual(2)
}
