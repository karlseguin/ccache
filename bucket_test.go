package ccache

import (
	. "github.com/karlseguin/expect"
	"testing"
	"time"
)

type BucketTests struct {
}

func Tests_Bucket(t *testing.T) {
	Expectify(new(BucketTests), t)
}

func (b *BucketTests) GetMissFromBucket() {
	bucket := testBucket()
	Expect(bucket.get("invalid")).To.Equal(nil)
}

func (b *BucketTests) GetHitFromBucket() {
	bucket := testBucket()
	item := bucket.get("power")
	assertValue(item, "9000")
}

func (b *BucketTests) DeleteItemFromBucket() {
	bucket := testBucket()
	bucket.delete("power")
	Expect(bucket.get("power")).To.Equal(nil)
}

func (b *BucketTests) SetsANewBucketItem() {
	bucket := testBucket()
	item, new := bucket.set("spice", TestValue("flow"), time.Minute)
	assertValue(item, "flow")
	item = bucket.get("spice")
	assertValue(item, "flow")
	Expect(new).To.Equal(true)
}

func (b *BucketTests) SetsAnExistingItem() {
	bucket := testBucket()
	item, new := bucket.set("power", TestValue("9002"), time.Minute)
	assertValue(item, "9002")
	item = bucket.get("power")
	assertValue(item, "9002")
	Expect(new).To.Equal(false)
}

func testBucket() *Bucket {
	b := &Bucket{lookup: make(map[string]*Item)}
	b.lookup["power"] = &Item{
		key:   "power",
		value: TestValue("9000"),
	}
	return b
}

func assertValue(item *Item, expected string) {
	value := item.value.(TestValue)
	Expect(value).To.Equal(TestValue(expected))
}

type TestValue string

func (v TestValue) Expires() time.Time {
	return time.Now()
}
