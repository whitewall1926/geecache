package lru

import (
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestGet(t *testing.T) {
	lru := New(int64(0), nil)

	lru.Add("key1", String("1234"))

	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1 failed")
	}

	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 expected")
	}
}

func TestRemoveOldest(t *testing.T) {
	lru := New(int64(16), nil)
	lru.Add("key1", String("1234"))
	lru.Add("key2", String("1234"))
	lru.Add("key3", String("1234"))

	if _, ok := lru.Get("key1"); ok {
		t.Fatalf("key1 should be evicted")
	}
	if v, ok := lru.Get("key2"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("key2 should remain")
	}
	if v, ok := lru.Get("key3"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("key3 should remain")
	}
}

func TestGetUpdatesRecency(t *testing.T) {
	lru := New(int64(16), nil)
	lru.Add("key1", String("1234"))
	lru.Add("key2", String("1234"))

	if _, ok := lru.Get("key1"); !ok {
		t.Fatalf("key1 should exist")
	}

	lru.Add("key3", String("1234"))

	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("key2 should be evicted after key1 is touched")
	}
	if _, ok := lru.Get("key1"); !ok {
		t.Fatalf("key1 should remain as most recently used")
	}
	if _, ok := lru.Get("key3"); !ok {
		t.Fatalf("key3 should exist")
	}
}

func TestOnEvicted(t *testing.T) {
	var callbackKey string
	var callbackVal string

	lru := New(int64(10), func(key string, value Value) {
		callbackKey = key
		callbackVal = string(value.(String))
	})

	lru.Add("k1", String("1234")) // 2 + 4 = 6
	lru.Add("k2", String("1234")) // 6 + 6 = 12, should evict k1

	if callbackKey != "k1" || callbackVal != "1234" {
		t.Fatalf("unexpected callback result: key=%s value=%s", callbackKey, callbackVal)
	}
}

func TestAddExistingKey(t *testing.T) {
	lru := New(int64(10), nil)
	lru.Add("k1", String("1234"))
	lru.Add("k1", String("12"))

	if v, ok := lru.Get("k1"); !ok || string(v.(String)) != "12" {
		t.Fatalf("k1 should be updated")
	}

	lru.Add("k2", String("1234"))
	if _, ok := lru.Get("k1"); !ok {
		t.Fatalf("k1 should remain after update with smaller size")
	}
}

func TestMaxBytesZeroMeansNoLimit(t *testing.T) {
	lru := New(int64(0), nil)
	for i := 0; i < 100; i++ {
		lru.Add(string(rune('a'+(i%26)))+string(rune('A'+(i%26))), String("123456"))
	}

	if lru.ll.Len() == 0 {
		t.Fatalf("cache should keep items when maxBytes is zero")
	}
}
