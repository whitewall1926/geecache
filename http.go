package geecache

import (
	"fmt"
	"geecache/consistenthash"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type HTTPPool struct {
	self        string
	basePath    string
	mu          sync.Mutex
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter
}

type httpGetter struct {
	baseURL string
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: "/_geecache/",
	}
}

func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(50, nil)

	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter)

	for _, v := range peers {
		p.httpGetters[v] = &httpGetter{
			baseURL: fmt.Sprintf("%v%v", v, p.basePath),
		}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 第一步：问大脑（哈希环），这个 key 归哪台机器管？
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		// 第二步：算出机器名（peer）后，再去通讯录（httpGetters）里查！
		p.Log("Pick peer %s", peer) // 可选日志
		return p.httpGetters[peer], true
	}
	return nil, false
}

func (p *httpGetter) Get(group string, key string) ([]byte, error) {
	path := fmt.Sprintf(
		"%v%v/%v",
		p.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)

	resp, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", resp.Status)
	}
	
	bytes, err := io.ReadAll(resp.Body)

	return bytes, err
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ok := strings.HasPrefix(r.URL.Path, p.basePath)
	if !ok {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	group_key := r.URL.Path[len(p.basePath):]
	parts := strings.SplitN(group_key, "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName, key := parts[0], parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}
