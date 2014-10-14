# CCache
CCache is an LRU Cache, written in Go, focused on supporting high concurrency.

Lock contention on the list is reduced by:

* Introducing a window which limits the frequency that an item can get promoted
* Using a buffered channel to queue promotions for a single worker
* Garbage collecting within the same thread as the worker

## Setup

First, download the project:

    go get github.com/karlseguin/ccache

## Configuration
Next, import and create a `ccache` instance:


    import (
      "github.com/karlseguin/ccache"
    )

    var cache = ccache.New(ccache.Configure())

`Configure` exposes a chainable API:

    var cache = ccache.New(ccache.Configure().MaxItems(1000).itemsToPrune(100))

The most likely configuration options to tweak are:

* `MaxItems(int)` - the maximum number of items to store in the cache (default: 5000)
* `GetsPerPromote(int)` - the number of times an item is fetched before we promote it. For large caches with long TTLs, it normally isn't necessary to promote an item after every fetch (default: 3)
* `ItemsToPrune(int)` - the number of items to prune when we hit `MaxItems`. Freeing up more than 1 slot at a time improved performance (default: 500)

Configurations that change the internals of the cache, which aren't as likely to need tweaking:

* `Buckets` - ccache shards its internal map to provide a greater amount of concurrency. The number of buckets is configurable (default: 16)
* `PromoteBuffer(int)` - the size of the buffer to use to queue promotions (default: 1024)
* `DeleteBuffer(int)` the size of the buffer to use to queue deletions (default: 1024)

## Usage

Once the cache is setup, you can  `Get`, `Set` and `Delete` items from it. A `Get` returns an `interface{}` which you'll want to cast back to the type of object you stored:

    item := cache.Get("user:4")
    if item == nil {
      //handle
    } else {
      user := item.(*User)
    }

`Set` expects the key, value and ttl:

    cache.Set("user:4", user, time.Minute * 10)

There's also a `Fetch` which mixes a `Get` and a `Set`:

    item, err := cache.Fetch("user:4", time.Minute * 10, func() (interface{}, error) {
      //code to fetch the data incase of a miss
      //should return the data to cache and the error, if any
    })

## Tracking
ccache supports a special tracking mode which is meant to be used in conjunction with other pieces of your code that maintains a long-lived reference to data.

When you configure your cache with `Track()`:

    cache = ccache.New(ccache.Configure().Track())

The items retrieved via `TrackingGet` will not be eligible for purge until `Release` is called on them:

    item := cache.TrackingGet("user:4")
    user := item.Value()   //will be nil if "user:4" didn't exist in the cache
    item.Release()  //can be called even if item.Value() returned nil

In practive, `Release` wouldn't be called until later, at some other place in your code.

There's a couple reason to use the tracking mode if other parts of your code also hold references to objects. First, if you're already going to hold a reference to these objects, there's really no reason not to have them in the cache - the memory is used up anyways.

More important, it helps ensure that you're code returns consistent data. With tracking, "user:4" might be purged, and a subsequent `Fetch` would reload the data. This can result in different versions of "user:4" being returned by different parts of your system.
