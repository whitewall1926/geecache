package consistenthash

import (
	"fmt"
	"math"
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	// 架构师的偷梁换柱：伪造一个极其简单的 Hash 函数
	// 传入 "6"，返回数字 6
	mockHash := func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	}

	// 初始化一致性哈希环，replicas (虚拟节点倍数) 设为 1 方便测试
	hash := New(1, mockHash)

	// 对应你的预设：入列 2, 4, 6
	hash.Add("6", "4", "2")

	// 准备你的测试用例字典：Key 是查询的数据，Value 是期望命中的机器
	testCases := map[string]string{
		"1": "2", // 场景 1：命中第一台机器 (1 -> 2)
		"3": "4", // 场景 2：命中第二台机器 (3 -> 4)
		"7": "2", // 场景 3：越界闭环，命中第一台机器 (7 -> 2)
	}

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("查询 %s 失败：期望命中 %s，实际命中 %s", k, v, hash.Get(k))
		}
	}

	// 场景 4：动态扩容，增加节点 "8"
	hash.Add("8")

	// 此时查询 "7"，期望目标从 "2" 变成了新机器 "8"
	if hash.Get("7") != "8" {
		t.Errorf("动态扩容测试失败：查询 7 期望命中 8，实际命中 %s", hash.Get("7"))
	}
}

func TestGet_ExtremeBoundaryWrapAround(t *testing.T) {
	// 构造一个可控 hash：把节点和查询键映射到 uint32 的边界附近。
	extremeHash := func(key []byte) uint32 {
		switch string(key) {
		case "0low":
			return 1
		case "0high":
			return math.MaxUint32 - 1
		case "belowLow":
			return 0
		case "between":
			return math.MaxUint32 / 2
		case "aboveHigh":
			return math.MaxUint32
		default:
			return 0
		}
	}

	hash := New(1, extremeHash)
	hash.Add("low", "high")

	if got := hash.Get("belowLow"); got != "low" {
		t.Fatalf("belowLow 命中错误：期望 low，实际 %s", got)
	}

	if got := hash.Get("between"); got != "high" {
		t.Fatalf("between 命中错误：期望 high，实际 %s", got)
	}

	if got := hash.Get("aboveHigh"); got != "low" {
		t.Fatalf("aboveHigh 回绕错误：期望 low，实际 %s", got)
	}
}

func TestDistributionWithReplicas(t *testing.T) {
	nodes := []string{"nodeA", "nodeB", "nodeC", "nodeD"}
	const totalKeys = 10000
	countFor := func(replicas int) map[string]int {
		hash := New(replicas, nil)
		hash.Add(nodes...)

		counts := make(map[string]int, len(nodes))
		for _, node := range nodes {
			counts[node] = 0
		}

		for i := 0; i < totalKeys; i++ {
			key := fmt.Sprintf("key-%d", i)
			node := hash.Get(key)
			counts[node]++
		}
		return counts
	}

	spread := func(counts map[string]int) int {
		minCount, maxCount := totalKeys, 0
		for _, node := range nodes {
			if counts[node] < minCount {
				minCount = counts[node]
			}
			if counts[node] > maxCount {
				maxCount = counts[node]
			}
		}
		return maxCount - minCount
	}

	countsWithoutReplicas := countFor(1)
	countsWithReplicas := countFor(50)

	for _, node := range nodes {
		if countsWithReplicas[node] == 0 {
			t.Fatalf("节点 %s 没有分配到任何 key，分布异常: %v", node, countsWithReplicas)
		}
	}

	if got, want := spread(countsWithReplicas), spread(countsWithoutReplicas); got >= want {
		t.Fatalf(
			"虚拟节点没有改善分布: replicas=50 spread=%d, replicas=1 spread=%d, with=%v, without=%v",
			got, want, countsWithReplicas, countsWithoutReplicas,
		)
	}
}
