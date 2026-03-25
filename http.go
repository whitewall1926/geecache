package geecache

import (
	"net/http"
	"strings"
)

type HTTPPool struct {
	self string
	basePath string
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self: self,
		basePath: "/_geecache/",
	}
}

func (p * HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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