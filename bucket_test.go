package ccache

import (
  "time"
  "testing"
  "github.com/karlseguin/gspec"
)

func TestGetMissFromBucket(t *testing.T) {
  bucket := testBucket()
  gspec.New(t).Expect(bucket.get("invalid")).ToBeNil()
}

func TestGetHitFromBucket(t *testing.T) {
  bucket := testBucket()
  item := bucket.get("power")
  assertValue(t, item, "9000")
}

func TestDeleteItemFromBucket(t *testing.T) {
  bucket := testBucket()
  bucket.delete("power")
  gspec.New(t).Expect(bucket.get("power")).ToBeNil()
}

func TestSetsANewBucketItem(t *testing.T) {
  bucket := testBucket()
  item := bucket.set("spice", TestValue("flow"), time.Minute)
  assertValue(t, item, "flow")
  item = bucket.get("spice")
  assertValue(t, item, "flow")
}

func TestSetsAnExistingItem(t *testing.T) {
  bucket := testBucket()
  item := bucket.set("power", TestValue("9002"), time.Minute)
  assertValue(t, item, "9002")
  item = bucket.get("power")
  assertValue(t, item, "9002")
}

func testBucket() *Bucket {
  b := &Bucket{lookup: make(map[string]*Item),}
  b.lookup["power"] = &Item{
    key: "power",
    value: TestValue("9000"),
  }
  return b
}

func assertValue(t *testing.T, item *Item, expected string) {
  value := item.value.(TestValue)
  gspec.New(t).Expect(value).ToEqual(TestValue(expected))
}

type TestValue string

func (v TestValue) Expires() time.Time {
  return time.Now()
}
