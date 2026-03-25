package geecache

import (
	"fmt"
	"sync"
)

type Group struct {
	name string
	mainCache  cache
	getter Getter
}

var mu sync.RWMutex
var groups = make(map[string]*Group)


func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter!")
	}
	mu.Lock()
	defer mu.Unlock()

	Group := &Group{
		name: name,
		mainCache: cache{cacheBytes: cacheBytes},
		getter: getter,

	}
	groups[name] = Group
	return Group
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	group, ok := groups[name]
	if !ok {
		return nil
	}
	return group
}

func (g *Group) Get(key string) (ByteView, error) {
	if len(key) == 0{
		return ByteView{}, fmt.Errorf("键不存在")
	}
	value, ok := g.mainCache.get(key)
	if ok {
		return value.(ByteView), nil
	}
	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	new_bytes := make([]byte, len(bytes))
	copy(new_bytes, bytes)
	new_bytes_byteview := ByteView{b: new_bytes}
	g.mainCache.add(key, new_bytes_byteview)
	return new_bytes_byteview, nil
}
