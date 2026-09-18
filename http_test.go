package geecache

import (
	"bytes"
	"fmt"
	"geecache/geecachepb"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/protobuf/proto"
)

// 为了测试，我们快速搞一个临时的 Getter
type testHTTPGetter struct{}

func (g *testHTTPGetter) Get(key string) ([]byte, error) {
	if key == "Tom" {
		return []byte("630"), nil
	}
	return nil, fmt.Errorf("not found")
}

func TestHTTPPool_ServeHTTP(t *testing.T) {
	// 1. 初始化我们要测试的军团和基站
	NewGroup("http_scores", 1024, &testHTTPGetter{})
	pool := NewHTTPPool("http://localhost:8000")

	// 2. 按当前协议构造 protobuf 请求体
	reqMsg := &geecachepb.Request{
		Group: "http_scores",
		Key:   "Tom",
	}
	body, err := proto.Marshal(reqMsg)
	if err != nil {
		t.Fatalf("序列化请求失败: %v", err)
	}
	req := httptest.NewRequest("POST", "http://localhost:8000/_geecache/", bytes.NewReader(body))

	// 3. 准备一个“录音机”，用来充当 ResponseWriter 接收返回值
	w := httptest.NewRecorder()

	pool.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望成功，返回错误码%v", w.Code)
	}

	var respMsg geecachepb.Response
	if err := proto.Unmarshal(w.Body.Bytes(), &respMsg); err != nil {
		t.Fatalf("反序列化响应失败: %v", err)
	}
	if got := string(respMsg.GetValue()); got != "630" {
		t.Fatalf("期望命中 Tom=630，获得一个错误的值Tom=%v", got)
	}
}

func TestHTTPPoolServeHTTPRejectsInvalidRequest(t *testing.T) {
	pool := NewHTTPPool("http://localhost:8000")
	req := httptest.NewRequest("POST", "http://localhost:8000/_geecache/", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	pool.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHTTPPoolServeHTTPRejectsUnknownGroup(t *testing.T) {
	request, err := proto.Marshal(&geecachepb.Request{Group: "missing", Key: "key"})
	if err != nil {
		t.Fatalf("serialize request: %v", err)
	}
	req := httptest.NewRequest("POST", "http://localhost:8000/_geecache/", bytes.NewReader(request))
	w := httptest.NewRecorder()

	NewHTTPPool("http://localhost:8000").ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHTTPGetterRejectsNonOKResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	getter := &httpGetter{baseURL: server.URL}
	if _, err := getter.Get("group", "key"); err == nil {
		t.Fatal("expected an error for a non-200 response")
	}
}
