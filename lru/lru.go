package lru

import (
	"container/list"
)

type Value interface {
	Len() int
}

type entry struct {
	key string
	value Value
}

type Cache struct {
	maxBytes int64
	nBytes int64
	ll *list.List
	kv map[string]*list.Element
	OnEvicted func(key string, value Value)
}


func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes: maxBytes,
		ll: list.New(),
		kv: make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string)(value Value, ok bool) {
	ele, ok := c.kv[key]
	if !ok {
		return nil, false
	}
	c.ll.MoveToFront(ele)
	kvEntry := ele.Value.(*entry)
	return kvEntry.value, ok
}

func (c * Cache) Removeoldest() {
	ele := c.ll.Back()
	if ele == nil {
		return 
	}

	kvEntry, ok := ele.Value.(*entry)
	if !ok {
		return 
	}
	c.nBytes -= int64(len(kvEntry.key)) + int64(kvEntry.value.Len())
	
	delete(c.kv, kvEntry.key)
	c.ll.Remove(ele)
	
	if c.OnEvicted != nil {
		c.OnEvicted(kvEntry.key, kvEntry.value)
	}

}

func (c * Cache) Add(key string, value Value) {

	ele, ok := c.kv[key]
	
	if !ok {
		node := &entry{key: key, value: value}
		c.ll.PushFront(node)
		c.nBytes += int64(node.value.Len()) + int64(len(node.key))
		c.kv[key] = c.ll.Front()
	} else {
		c.ll.MoveToFront(ele)
		kvEntry := ele.Value.(*entry)
		c.nBytes -= int64(kvEntry.value.Len())
		kvEntry.value = value
		c.nBytes += int64(kvEntry.value.Len())
	}

	for c.ll.Back() != nil && c.maxBytes != 0 &&  c.nBytes > c.maxBytes {
			c.Removeoldest()
		}
}