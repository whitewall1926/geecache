package geecache

import (
    "fmt"
    // "io"
    "net/http"
    "net/http/httptest"
    "testing"
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

    // 2. 伪造一个 HTTP 请求 (请求 URL 是 /_geecache/http_scores/Tom)
    req := httptest.NewRequest("GET", "http://localhost:8000/_geecache/http_scores/Tom", nil)
    
    // 3. 准备一个“录音机”，用来充当 ResponseWriter 接收返回值
    w := httptest.NewRecorder()

    // ==========================================
    // 你的任务开始：TODO
    // ==========================================
    
    // 任务 A：让基站（pool）去处理这个伪造的请求 (w, req)
    // 提示：调用 pool 的什么方法？
    pool.ServeHTTP(w, req)

    // 任务 B：验证 HTTP 状态码是否正确
    // 录音机 w 有一个 Code 字段，记录了返回的状态码。
    // 如果 w.Code 不等于 http.StatusOK (也就是 200)，请用 t.Fatalf 报错。
    if w.Code != http.StatusOK {
		t.Fatalf("期望成功，返回错误码%v", w.Code)
	}

    // 任务 C：验证返回的数据是否正确
    // 录音机 w 的 Body 字段记录了写入的数据流，可以使用 w.Body.Bytes() 取出 []byte。
    // 将它转成 string 后，判断是否等于 "630"，如果不等于，用 t.Fatalf 报错。
	bytes := w.Body.Bytes()
	str :=string(bytes)
	if str != "630" {
		t.Fatalf("期望命中 Tom=630，获得一个错误的值Tom=%v", str)
	}

    // ==========================================
    // 你的任务结束
    // ==========================================
}