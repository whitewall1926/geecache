package geecache

import (
	"fmt"
	"geecache/singleflight"
	"log"
	"sync"
)

type Group struct {
	name string
	mainCache  cache
	getter Getter

	peers PeerPicker

	singleFlight  *singleflight.Group
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
		singleFlight: singleflight.NewGroup(),
	}
	groups[name] = Group
	return Group
}

func (g * Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}


func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
    // 调用 httpGetter.Get() 发起网络请求
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	new_bytes := make([]byte, len(bytes))
	copy(new_bytes, bytes)
	return ByteView{b: new_bytes}, nil
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


func (g *Group) load(key string) (value ByteView, err error) {
	// 看看我们有没有装备对讲机 (调度中心)

	v, err := g.singleFlight.Do(key, func() (interface{}, error) {
		if g.peers != nil {
		// 问调度中心，这个 key 归哪个兄弟管？
		if peer, ok := g.peers.PickPeer(key); ok {
			// 归兄弟管！去兄弟那里拿！
			
			if value, err := g.getFromPeer(peer, key); err == nil {
				log.Printf("[GeeCache] 从远端节点获取 %s 成功", key)
				return value, nil
			}
			// 如果兄弟宕机了，或者网络超时了，打印个日志，继续往下走，自己查本地
			log.Println("[GeeCache] Failed to get from peer", err)
		}
	}
		return  g.getLocally(key)
	})

	if err == nil{
		byteView := v.(ByteView)
		return byteView, nil
	}
	return 
}


func (g *Group) getLocally(key string) (ByteView, error) {
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
