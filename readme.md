# CCache
CCache is an LRU Cache, written in Go, focused on supporting high concurrency.

Lock contention on the list is reduced by:

* Introducing a window which limits the frequency that an item can get promoted
* Using a buffered channel to queue promotions for a single worker
* Garbage collecting within the same thread as the worker

