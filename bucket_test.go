package ccache

import (
	"testing"
	"time"

	. "github.com/karlseguin/expect"
)

type BucketTests struct {
}

func Test_Bucket(t *testing.T) {
	Expectify(new(BucketTests), t)
}

func (_ *BucketTests) GetMissFromBucket() {
	bucket := testBucket()
	Expect(bucket.get("invalid")).To.Equal(nil)
}

func (_ *BucketTests) GetHitFromBucket() {
	bucket := testBucket()
	item := bucket.get("power")
	assertValue(item, "9000")
}

func (_ *BucketTests) DeleteItemFromBucket() {
	bucket := testBucket()
	bucket.delete("power")
	Expect(bucket.get("power")).To.Equal(nil)
}

func (_ *BucketTests) SetsANewBucketItem() {
	bucket := testBucket()
	item, existing := bucket.set("spice", "flow", time.Minute, false)
	assertValue(item, "flow")
	item = bucket.get("spice")
	assertValue(item, "flow")
	Expect(existing).To.Equal(nil)
}

func (_ *BucketTests) SetsAnExistingItem() {
	bucket := testBucket()
	item, existing := bucket.set("power", "9001", time.Minute, false)
	assertValue(item, "9001")
	item = bucket.get("power")
	assertValue(item, "9001")
	assertValue(existing, "9000")
}

func testBucket() *bucket[string] {
	b := &bucket[string]{lookup: make(map[string]*Item[string])}
	b.lookup["power"] = &Item[string]{
		key:   "power",
		value: "9000",
	}
	return b
}

func assertValue(item *Item[string], expected string) {
	Expect(item.value).To.Equal(expected)
}
