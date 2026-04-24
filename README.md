
# geecache

一个用 Go 实现的简化版分布式缓存项目，包含本地 LRU 缓存、一致性哈希节点选择、singleflight 并发请求合并，以及基于 HTTP + protobuf 的节点间通信。

## 项目目标

- 理解缓存命中、缓存回源和缓存填充的基本流程
- 理解一致性哈希如何将 key 路由到节点
- 理解 singleflight 如何合并同一 key 的并发请求
- 用尽量小的代码把本地缓存和分布式节点协作串起来

## 查询流程

查询一个 key 时，主流程如下：

1. 调用 `Group.Get(key)`
2. 先查本地缓存 `mainCache`
3. 本地未命中时，进入 `singleFlight.Do(key, fn)`，合并同一 key 的并发请求
4. 如果配置了 `peers`，通过一致性哈希选择负责节点
5. 如果 key 属于远端节点，则通过 HTTP + protobuf 发起远程获取
6. 如果没有远端节点、远端失败，或者 key 本就归当前节点，则调用本地 `getter`
7. 本地 `getter` 获取成功后，写回本地缓存并返回

对应代码：

- [geecache.go](/home/yxf/geecache/geecache.go:68)
- [http.go](/home/yxf/geecache/http.go:55)
- [consistenthash/consistenthash.go](/home/yxf/geecache/consistenthash/consistenthash.go:38)
- [singleflight/singleflight.go](/home/yxf/geecache/singleflight/singleflight.go:22)

## 代码结构

- `geecache.go`: `Group` 和缓存查询主流程
- `local_cache.go`: 本地缓存抽象 `LocalCache`
- `cache_lru.go`: 默认本地缓存实现，基于 `lru.Cache`
- `http.go`: 节点选择、远程获取、HTTP 服务入口
- `byteview.go`: 不可变缓存值封装
- `peers.go`: `PeerPicker` / `PeerGetter` 抽象
- `lru/`: LRU 缓存实现
- `consistenthash/`: 一致性哈希实现
- `singleflight/`: 并发请求合并实现
- `geecachepb/`: HTTP 节点通信使用的 protobuf 消息
- `cmd/main.go`: 示例启动入口

## 设计说明

### 1. `Group` 依赖抽象而不是具体缓存实现

`Group` 现在依赖的是 `LocalCache` 接口，而不是直接依赖某个具体 `cache` 类型。这样做的目的：

- 让 `Group` 只负责查询流程
- 把 LRU 细节隔离在默认实现里
- 后续如果要替换成别的本地缓存策略，不需要改 `Group`

### 2. HTTP 通信协议

当前节点间通信使用：

- HTTP
- protobuf request/response body

测试已经按这个协议对齐，不再使用 URL path 直接传 `group/key`。

### 3. 一致性哈希测试策略

一致性哈希这部分除了基础命中测试，还增加了“虚拟节点是否改善分布”的验证。这里不强行要求绝对均匀，而是验证 `replicas=50` 相比 `replicas=1` 的分布离散度更小，更符合当前实现的真实目标。

## 运行

启动一个节点：

```bash
go run ./cmd -port=8001
```

也可以分别启动：

```bash
go run ./cmd -port=8001
go run ./cmd -port=8002
go run ./cmd -port=8003
```

`cmd/main.go` 里内置了一个简单的示例数据源：

- `Tom -> 630`
- `Jack -> 589`
- `Sam -> 567`

## 测试

运行全部测试：

```bash
go test ./...
```

单独运行一致性哈希测试：

```bash
go test ./consistenthash
```
