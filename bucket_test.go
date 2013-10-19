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

func TestRemovesItemFromBucket(t *testing.T) {
  bucket := testBucket()
  bucket.remove("power")
  gspec.New(t).Expect(bucket.get("power")).ToBeNil()
}

func TestSetsANewBucketItem(t *testing.T) {
  bucket := testBucket()
  item := bucket.set("spice", newTestValue("flow"))
  assertValue(t, item, "flow")
  item = bucket.get("spice")
  assertValue(t, item, "flow")
}

func TestSetsAnExistingItem(t *testing.T) {
  bucket := testBucket()
  item := bucket.set("power", newTestValue("9002"))
  assertValue(t, item, "9002")
  item = bucket.get("power")
  assertValue(t, item, "9002")
}

func testBucket() *Bucket {
  b := &Bucket{lookup: make(map[string]*Item),}
  b.lookup["power"] = &Item{
    key: "power",
    value: newTestValue("9000"),
  }
  return b
}

func assertValue(t *testing.T, item *Item, expected string) {
  value := item.value.(*TestValue)
  gspec.New(t).Expect(value.v).ToEqual(expected)
}

type TestValue struct {
  v string
}

func newTestValue(v string) *TestValue {
  return &TestValue{v: v,}
}

func (v *TestValue) Expires() time.Time {
  return time.Now()
}
