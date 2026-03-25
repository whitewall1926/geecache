package geecache

import (
    "fmt"
    "testing"
)

// 模拟的底层数据库
var db = map[string]string{
    "Tom":  "630",
    "Jack": "589",
    "Sam":  "567",
}

type mockGetter struct {
	called int
}

func (c *mockGetter)Get(key string) ([]byte, error) {
	c.called ++
	fmt.Printf("第%v次 从数据库中调用的\n", c.called)
	
	v, ok := db[key]
	if ok {
		return []byte(v), nil
	}
	return nil, fmt.Errorf("查找失败，数据库不存在此数据")
}

func TestGet(t *testing.T) {
	group := NewGroup("scores", 50, &mockGetter{})
	bytes, err := group.Get("Tom")

	if err != nil {
		t.Fatalf("期望成功的，但是报错了%v", err)
	}
	str := bytes.String()
	if str != "630" {
		t.Fatalf("期望通过，但是没有通过")
	}
	
	nbytes, _ := group.Get("Tom")
	new_str := nbytes.String()
	if new_str != "630" {
		t.Fatalf("期望通过，但是失败了")
	}

	 group.Get("Jack")

	  group.Get("Tom")
	  group.Get("Jack")
}