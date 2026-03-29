package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Map struct {
	hash     func(data []byte) uint32 // 你的 hash 函数
	replicas int                      // 比如你写的 3
	keys     []int                    // 你的 hashReplicas
	hashMap  map[int]string           // 你的 real_nodes
}

func New(replicas int , fn func(data []byte) uint32) *Map {
	if fn == nil {
		fn = crc32.ChecksumIEEE
	}
	return  &Map{
		hash: fn,
		replicas: replicas,
		hashMap: make(map[int]string),
	}
}

func (m *Map) Add(nodes ...string) {
	for _, name := range nodes {
		for i := 0; i < m.replicas; i++ {
			hashValue := int(m.hash([]byte(strconv.Itoa(i) + name)))
			m.keys = append(m.keys, hashValue)
			m.hashMap[hashValue] = name
		}
	}
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hashValue := m.hash([]byte(key))

	l, r := 0, len(m.keys)
	for l < r {
		mid := (l + r) / 2
		if m.keys[mid] >= int(hashValue) {
			r = mid
		} else {
			l = mid + 1
		}
	}
	if l == len(m.keys){
		return m.hashMap[m.keys[0]]
	}
	return m.hashMap[m.keys[l]]
}