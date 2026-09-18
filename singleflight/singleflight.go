package singleflight

import (
	"sync"
)

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

var (
	groupMu       sync.Mutex
	groupInstance *Group
)

func NewGroup() *Group {
	groupMu.Lock()
	defer groupMu.Unlock()

	if groupInstance == nil {
		groupInstance = &Group{
			m: make(map[string]*call),
		}
	}

	return groupInstance
}

// func NewGroup() *Group{

//		return &Group{
//			m: make(map[string]*call),
//		}
//	}
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 你来填
	//有人在处理->wait->从call里拿结果返回
	//没人处理->创建call->Add(1)->执行fn->Done->返回结果
	g.mu.Lock()
	c, ok := g.m[key]
	if !ok {
		c = new(call)
		c.wg.Add(1)
		g.m[key] = c
		g.mu.Unlock()
		c.val, c.err = fn()

		c.wg.Done()

		g.mu.Lock()
		delete(g.m, key)
		g.mu.Unlock()

		return c.val, c.err
	} else {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

}
