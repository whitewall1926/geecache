package main

import (
	"flag"
	"fmt"
	"geecache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// addr: 当前节点的地址，比如 "http://localhost:8001"
// addrs: 集群里所有节点的地址列表
// gee: 我们刚才在 main 里创建的大管家
func startCacheServer(addr string, addrs []string, gee *geecache.Group, port int) {
	// 1. 创建一个 HTTPPool（对讲机基站）
	// 2. 把所有节点的地址列表 (addrs) 录入到这个基站的通讯录（哈希环）里
	// 3. 将基站绑定到大管家 (gee) 身上，这样大管家找不到数据时就能用对讲机问别人了
	// 4. 启动 Go 语言内置的 HTTP 网络服务，监听本地端口
	pool := geecache.NewHTTPPool(addr)

	pool.Set(addrs...)

	gee.RegisterPeers(pool)
	log.Println("geecache is running at", addr)
	// Listen on all network interfaces so clients on other machines can connect.
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), pool))
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8001, "Geecache serve port")
	flag.Parse()

	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gee := geecache.NewGroup("scores", 2048, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)

			v, ok := db[key]
			if !ok {
				return nil, fmt.Errorf("数据库不存在对应的键：%s", key)
			} else {
				return []byte(v), nil
			}
		}))

	currentAddr := addrMap[port]
	startCacheServer(currentAddr, addrs, gee, port)
}
