package geecache

import (
	"geecache/lru"
	"sync"
)

type lruCache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

func newLocalCache(cacheBytes int64) LocalCache {
	return &lruCache{cacheBytes: cacheBytes}
}

func (c *lruCache) Add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}

	c.lru.Add(key, value)
}

func (c *lruCache) Get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		return ByteView{}, false
	}
	v, ok := c.lru.Get(key)
	if !ok {
		return ByteView{}, false
	}
	return v.(ByteView), true
}
