package ccache

import (
	. "github.com/karlseguin/expect"
	"strconv"
	"testing"
	"time"
)

type CacheTests struct{}

func Test_Cache(t *testing.T) {
	Expectify(new(CacheTests), t)
}

func (c *CacheTests) DeletesAValue() {
	cache := New(Configure())
	cache.Set("spice", "flow", time.Minute)
	cache.Set("worm", "sand", time.Minute)
	cache.Delete("spice")
	Expect(cache.Get("spice")).To.Equal(nil)
	Expect(cache.Get("worm").(string)).To.Equal("sand")
}

func (c *CacheTests) GCsTheOldestItems() {
	cache := New(Configure().ItemsToPrune(10))
	for i := 0; i < 500; i++ {
		cache.Set(strconv.Itoa(i), i, time.Minute)
	}
	//let the items get promoted (and added to our list)
	time.Sleep(time.Millisecond * 10)
	cache.gc()
	Expect(cache.Get("9")).To.Equal(nil)
	Expect(cache.Get("10").(int)).To.Equal(10)
}

func (c *CacheTests) PromotedItemsDontGetPruned() {
	cache := New(Configure().ItemsToPrune(10).GetsPerPromote(1))
	for i := 0; i < 500; i++ {
		cache.Set(strconv.Itoa(i), i, time.Minute)
	}
	time.Sleep(time.Millisecond * 10) //run the worker once to init the list
	cache.Get("9")
	time.Sleep(time.Millisecond * 10)
	cache.gc()
	Expect(cache.Get("9").(int)).To.Equal(9)
	Expect(cache.Get("10")).To.Equal(nil)
	Expect(cache.Get("11").(int)).To.Equal(11)
}

func (c *CacheTests) TrackerDoesNotCleanupHeldInstance() {
	cache := New(Configure().ItemsToPrune(10).Track())
	for i := 0; i < 10; i++ {
		cache.Set(strconv.Itoa(i), i, time.Minute)
	}
	item := cache.TrackingGet("0")
	time.Sleep(time.Millisecond * 10)
	cache.gc()
	Expect(cache.Get("0").(int)).To.Equal(0)
	Expect(cache.Get("1")).To.Equal(nil)
	item.Release()
	cache.gc()
	Expect(cache.Get("0")).To.Equal(nil)
}

func (c *CacheTests) RemovesOldestItemWhenFull() {
	cache := New(Configure().MaxItems(5).ItemsToPrune(1))
	for i := 0; i < 7; i++ {
		cache.Set(strconv.Itoa(i), i, time.Minute)
	}
	time.Sleep(time.Millisecond * 10)
	Expect(cache.Get("0")).To.Equal(nil)
	Expect(cache.Get("1")).To.Equal(nil)
	Expect(cache.Get("2").(int)).To.Equal(2)
}
