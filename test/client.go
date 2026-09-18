package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"geecache/geecachepb"

	"google.golang.org/protobuf/proto"
)

func main() {
	var target string
	var group string
	var hotKey string
	var keysCSV string
	var mode string
	var concurrency int
	var totalRequests int
	var hotRatio int

	flag.StringVar(&target, "target", "http://localhost:8001/_geecache/", "target endpoint")
	flag.StringVar(&group, "group", "scores", "group name")
	flag.StringVar(&hotKey, "hot-key", "Tom", "hot key when mode=hot|mixed")
	flag.StringVar(&keysCSV, "keys", "Tom,Jack,Sam", "candidate keys, comma-separated")
	flag.StringVar(&mode, "mode", "mixed", "traffic mode: hot|random|mixed")
	flag.IntVar(&concurrency, "c", 200, "concurrency")
	flag.IntVar(&totalRequests, "n", 20000, "total requests")
	flag.IntVar(&hotRatio, "hot-ratio", 80, "hot key ratio for mixed mode (0-100)")
	flag.Parse()

	keys := parseKeys(keysCSV)
	if len(keys) == 0 {
		panic("-keys must contain at least one key")
	}
	if hotRatio < 0 || hotRatio > 100 {
		panic("-hot-ratio must be between 0 and 100")
	}

	transport := &http.Transport{
		MaxIdleConns:        2000,
		MaxIdleConnsPerHost: 2000,
		IdleConnTimeout:     90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	var success int64
	var failed int64
	latencies := make([]time.Duration, 0, totalRequests)
	var latMu sync.Mutex

	statusCount := make(map[int]int64)
	var statusMu sync.Mutex

	jobs := make(chan int, totalRequests)
	var wg sync.WaitGroup

	start := time.Now()

	worker := func(workerID int) {
		defer wg.Done()
		rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)*97))

		for range jobs {
			t0 := time.Now()
			key := chooseKey(rng, mode, hotKey, keys, hotRatio)

			req := &geecachepb.Request{Group: group, Key: key}
			payload, err := proto.Marshal(req)
			if err != nil {
				atomic.AddInt64(&failed, 1)
				continue
			}

			resp, err := client.Post(target, "application/octet-stream", bytes.NewReader(payload))
			if err != nil {
				atomic.AddInt64(&failed, 1)
				continue
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()

			statusMu.Lock()
			statusCount[resp.StatusCode]++
			statusMu.Unlock()

			if err != nil || resp.StatusCode != http.StatusOK {
				atomic.AddInt64(&failed, 1)
				continue
			}

			var out geecachepb.Response
			if err := proto.Unmarshal(body, &out); err != nil {
				atomic.AddInt64(&failed, 1)
				continue
			}

			_ = out.GetValue()
			atomic.AddInt64(&success, 1)

			cost := time.Since(t0)
			latMu.Lock()
			latencies = append(latencies, cost)
			latMu.Unlock()
		}
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker(i)
	}

	for i := 0; i < totalRequests; i++ {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	totalCost := time.Since(start)

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	p50 := percentile(latencies, 50)
	p95 := percentile(latencies, 95)
	p99 := percentile(latencies, 99)

	qps := float64(totalRequests) / totalCost.Seconds()

	fmt.Println("=== Result ===")
	fmt.Println("Target        :", target)
	fmt.Println("Group         :", group)
	fmt.Println("Mode          :", mode)
	fmt.Println("Hot Key       :", hotKey)
	fmt.Println("Hot Ratio     :", hotRatio)
	fmt.Println("Keys          :", strings.Join(keys, ","))
	fmt.Println("Total Requests:", totalRequests)
	fmt.Println("Concurrency   :", concurrency)
	fmt.Println("Success       :", success)
	fmt.Println("Failed        :", failed)
	fmt.Printf("Total Time    : %v\n", totalCost)
	fmt.Printf("QPS           : %.2f\n", qps)
	fmt.Printf("P50 Latency   : %v\n", p50)
	fmt.Printf("P95 Latency   : %v\n", p95)
	fmt.Printf("P99 Latency   : %v\n", p99)
	fmt.Println("Status Count  :", statusCount)
}

func parseKeys(keysCSV string) []string {
	parts := strings.Split(keysCSV, ",")
	keys := make([]string, 0, len(parts))
	for _, p := range parts {
		k := strings.TrimSpace(p)
		if k != "" {
			keys = append(keys, k)
		}
	}
	return keys
}

func chooseKey(rng *rand.Rand, mode, hotKey string, keys []string, hotRatio int) string {
	switch mode {
	case "hot":
		return hotKey
	case "random":
		return keys[rng.Intn(len(keys))]
	case "mixed":
		if rng.Intn(100) < hotRatio {
			return hotKey
		}
		return keys[rng.Intn(len(keys))]
	default:
		return hotKey
	}
}

func percentile(arr []time.Duration, p int) time.Duration {
	if len(arr) == 0 {
		return 0
	}
	sort.Slice(arr, func(i, j int) bool { return arr[i] < arr[j] })
	idx := (len(arr) - 1) * p / 100
	return arr[idx]
}
