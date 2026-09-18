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

func (c *mockGetter) Get(key string) ([]byte, error) {
	c.called++
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

	for _, key := range []string{"Jack", "Tom", "Jack"} {
		if _, err := group.Get(key); err != nil {
			t.Fatalf("期望成功的，但是报错了%v", err)
		}
	}
}

func TestGroupGetRejectsEmptyKey(t *testing.T) {
	getterCalls := 0
	group := NewGroup("empty-key", 50, GetterFunc(func(string) ([]byte, error) {
		getterCalls++
		return []byte("value"), nil
	}))

	if _, err := group.Get(""); err == nil {
		t.Fatal("expected an error for an empty key")
	}
	if getterCalls != 0 {
		t.Fatalf("getter should not be called for an empty key, got %d calls", getterCalls)
	}
}

func TestGroupGetDoesNotCacheGetterError(t *testing.T) {
	getterCalls := 0
	group := NewGroup("getter-error", 50, GetterFunc(func(string) ([]byte, error) {
		getterCalls++
		return nil, fmt.Errorf("backend unavailable")
	}))

	for i := 0; i < 2; i++ {
		if _, err := group.Get("key"); err == nil {
			t.Fatal("expected getter error")
		}
	}
	if getterCalls != 2 {
		t.Fatalf("failed values should not be cached, got %d getter calls", getterCalls)
	}
}

type testPeerPicker struct {
	peer PeerGetter
	ok   bool
}

func (p *testPeerPicker) PickPeer(string) (PeerGetter, bool) {
	return p.peer, p.ok
}

type testPeerGetter struct {
	value []byte
	err   error
}

func (p *testPeerGetter) Get(string, string) ([]byte, error) {
	return p.value, p.err
}

func TestGroupGetFromPeer(t *testing.T) {
	getterCalls := 0
	group := NewGroup("peer-success", 50, GetterFunc(func(string) ([]byte, error) {
		getterCalls++
		return []byte("local"), nil
	}))
	group.RegisterPeers(&testPeerPicker{
		peer: &testPeerGetter{value: []byte("remote")},
		ok:   true,
	})

	value, err := group.Get("key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := value.String(); got != "remote" {
		t.Fatalf("expected remote value, got %q", got)
	}
	if getterCalls != 0 {
		t.Fatalf("local getter should not be called, got %d calls", getterCalls)
	}
}

func TestGroupGetFallsBackToLocalOnPeerError(t *testing.T) {
	group := NewGroup("peer-error", 50, GetterFunc(func(string) ([]byte, error) {
		return []byte("local"), nil
	}))
	group.RegisterPeers(&testPeerPicker{
		peer: &testPeerGetter{err: fmt.Errorf("peer unavailable")},
		ok:   true,
	})

	value, err := group.Get("key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := value.String(); got != "local" {
		t.Fatalf("expected local fallback value, got %q", got)
	}
}
