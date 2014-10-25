package ccache

import (
	. "github.com/karlseguin/expect"
	"testing"
	"time"
	"strconv"
)

type LayeredCacheTests struct{}

func Test_LayeredCache(t *testing.T) {
	Expectify(new(LayeredCacheTests), t)
}

func (l *LayeredCacheTests) GetsANonExistantValue() {
	cache := newLayered()
	Expect(cache.Get("spice", "flow")).To.Equal(nil)
}

func (l *LayeredCacheTests) SetANewValue() {
	cache := newLayered()
	cache.Set("spice", "flow", "a value", time.Minute)
	Expect(cache.Get("spice", "flow").(string)).To.Equal("a value")
	Expect(cache.Get("spice", "stop")).To.Equal(nil)
}

func (l *LayeredCacheTests) SetsMultipleValueWithinTheSameLayer() {
	cache := newLayered()
	cache.Set("spice", "flow", "value-a", time.Minute)
	cache.Set("spice", "must", "value-b", time.Minute)
	cache.Set("leto", "sister", "ghanima", time.Minute)
	Expect(cache.Get("spice", "flow").(string)).To.Equal("value-a")
	Expect(cache.Get("spice", "must").(string)).To.Equal("value-b")
	Expect(cache.Get("spice", "worm")).To.Equal(nil)

	Expect(cache.Get("leto", "sister").(string)).To.Equal("ghanima")
	Expect(cache.Get("leto", "brother")).To.Equal(nil)
	Expect(cache.Get("baron", "friend")).To.Equal(nil)
}

func (l *LayeredCacheTests) DeletesAValue() {
	cache := newLayered()
	cache.Set("spice", "flow", "value-a", time.Minute)
	cache.Set("spice", "must", "value-b", time.Minute)
	cache.Set("leto", "sister", "ghanima", time.Minute)
	cache.Delete("spice", "flow")
	Expect(cache.Get("spice", "flow")).To.Equal(nil)
	Expect(cache.Get("spice", "must").(string)).To.Equal("value-b")
	Expect(cache.Get("spice", "worm")).To.Equal(nil)
	Expect(cache.Get("leto", "sister").(string)).To.Equal("ghanima")
}


func (l *LayeredCacheTests) DeletesALayer() {
	cache := newLayered()
	cache.Set("spice", "flow", "value-a", time.Minute)
	cache.Set("spice", "must", "value-b", time.Minute)
	cache.Set("leto", "sister", "ghanima", time.Minute)
	cache.DeleteAll("spice")
	Expect(cache.Get("spice", "flow")).To.Equal(nil)
	Expect(cache.Get("spice", "must")).To.Equal(nil)
	Expect(cache.Get("spice", "worm")).To.Equal(nil)
	Expect(cache.Get("leto", "sister").(string)).To.Equal("ghanima")
}





func (c *LayeredCacheTests) GCsTheOldestItems() {
	cache := Layered(Configure().ItemsToPrune(10))
	cache.Set("xx", "a", 23, time.Minute)
	for i := 0; i < 500; i++ {
		cache.Set(strconv.Itoa(i), "a", i, time.Minute)
	}
	cache.Set("xx", "b", 9001, time.Minute)
	//let the items get promoted (and added to our list)
	time.Sleep(time.Millisecond * 10)
	cache.gc()
	Expect(cache.Get("xx", "a")).To.Equal(nil)
	Expect(cache.Get("xx", "b").(int)).To.Equal(9001)
	Expect(cache.Get("8", "a")).To.Equal(nil)
	Expect(cache.Get("9", "a")).To.Equal(9)
	Expect(cache.Get("10", "a").(int)).To.Equal(10)
}

func (c *LayeredCacheTests) PromotedItemsDontGetPruned() {
	cache := Layered(Configure().ItemsToPrune(10).GetsPerPromote(1))
	for i := 0; i < 500; i++ {
		cache.Set(strconv.Itoa(i), "a", i, time.Minute)
	}
	time.Sleep(time.Millisecond * 10) //run the worker once to init the list
	cache.Get("9", "a")
	time.Sleep(time.Millisecond * 10)
	cache.gc()
	Expect(cache.Get("9", "a").(int)).To.Equal(9)
	Expect(cache.Get("10", "a")).To.Equal(nil)
	Expect(cache.Get("11", "a").(int)).To.Equal(11)
}

func (c *LayeredCacheTests) TrackerDoesNotCleanupHeldInstance() {
	cache := Layered(Configure().ItemsToPrune(10).Track())
	for i := 0; i < 10; i++ {
		cache.Set(strconv.Itoa(i), "a", i, time.Minute)
	}
	item := cache.TrackingGet("0", "a")
	time.Sleep(time.Millisecond * 10)
	cache.gc()
	Expect(cache.Get("0", "a").(int)).To.Equal(0)
	Expect(cache.Get("1", "a")).To.Equal(nil)
	item.Release()
	cache.gc()
	Expect(cache.Get("0", "a")).To.Equal(nil)
}

func (c *LayeredCacheTests) RemovesOldestItemWhenFull() {
	cache := Layered(Configure().MaxItems(5).ItemsToPrune(1))
	cache.Set("xx", "a", 23, time.Minute)
	for i := 0; i < 7; i++ {
		cache.Set(strconv.Itoa(i), "a", i, time.Minute)
	}
	cache.Set("xx", "b", 9001, time.Minute)
	time.Sleep(time.Millisecond * 10)
	Expect(cache.Get("xx", "a")).To.Equal(nil)
	Expect(cache.Get("0", "a")).To.Equal(nil)
	Expect(cache.Get("1", "a")).To.Equal(nil)
	Expect(cache.Get("2", "a")).To.Equal(nil)
	Expect(cache.Get("3", "a")).To.Equal(3)
	Expect(cache.Get("xx", "b")).To.Equal(9001)
}

func newLayered() *LayeredCache {
	return Layered(Configure())
}
