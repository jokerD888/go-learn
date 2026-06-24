// === Lesson 12：HTTP 服务 — net/http、JSON API、middleware ===
// 目标：用标准库构建一个 REST API，理解 handler/middleware 模式。

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// ── 数据模型 ──

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

// ── 内存存储（替代数据库，聚焦 HTTP 本身） ──
type TaskStore struct {
	tasks  []Task
	nextID int
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks: []Task{
			{ID: 1, Title: "学 Go", Done: true, CreatedAt: time.Now()},
			{ID: 2, Title: "写 HTTP 服务", Done: false, CreatedAt: time.Now()},
		},
		nextID: 3,
	}
}

func main() {
	store := NewTaskStore()

	// ── 方式 1：http.HandleFunc（全局默认 mux） ──
	// 简单场景够用。并发不安全（默认 mux 是全局变量），生产用自定义 mux。
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	// ── 方式 2：http.NewServeMux（Go 1.22+ 路由增强） ──
	// 自定义 mux，推荐。Go 1.22 后支持 METHOD 和路径变量。
	mux := http.NewServeMux()

	// GET /tasks —— 列表
	mux.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, store.tasks)
	})

	// POST /tasks —— 创建
	mux.HandleFunc("POST /tasks", func(w http.ResponseWriter, r *http.Request) {
		var task Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的 JSON"})
			return
		}
		task.ID = store.nextID
		store.nextID++
		task.CreatedAt = time.Now()
		store.tasks = append(store.tasks, task)
		writeJSON(w, http.StatusCreated, task)
	})

	// ── 中间件：包装 handler，添加横切逻辑 ──
	// Go 的中间件是 func(http.Handler) http.Handler —— 接受一个 handler，返回一个 handler。

	loggedMux := loggingMiddleware(mux)        // 日志
	loggedMux = corsMiddleware(loggedMux)      // CORS
	loggedMux = recoverMiddleware(loggedMux)   // panic 恢复

	// ── 启动服务 ──
	addr := ":8080"
	fmt.Printf("HTTP 服务启动: http://localhost%s\n", addr)
	fmt.Println("接口:")
	fmt.Println("  GET  /tasks   — 列出所有任务")
	fmt.Println("  POST /tasks   — 创建新任务")
	fmt.Println("  GET  /health  — 健康检查")

	// ListenAndServe 会阻塞直到出错或关闭。
	// ponytail: 生产用 gracehttp/shutdown，这里聚焦学习 HTTP 核心。
	if err := http.ListenAndServe(addr, loggedMux); err != nil {
		log.Fatal(err)
	}
}

// ── 工具函数 ──

// writeJSON 统一 JSON 响应，避免到处手写 header + encode。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) // Encode 比 Marshal+Write 少一次 []byte 分配
}

// ── 中间件实现 ──

// loggingMiddleware 记录每个请求的方法、路径和耗时。
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r) // 调用下一个 handler
		fmt.Printf("[%s] %s %s (%v)\n", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}

// corsMiddleware 允许跨域请求。
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		// 浏览器会先发 OPTIONS 预检请求，直接返回 204。
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// recoverMiddleware 捕获 handler 中的 panic，防止整个服务崩溃。
// recover 只能在 defer 中生效，且只捕获当前 goroutine 的 panic。
// panic 后同级代码跳过，defer 执行完后 goroutine 退出。
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v", err)
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "内部错误"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ── 生产化 checklist ──
// 1. 用 *http.ServeMux 替代 DefaultServeMux（避免全局状态）。
// 2. 中间件用 func(http.Handler) http.Handler 模式，可组合、可测试。
// 3. 永远设置 Content-Type header，客户端不猜测响应类型。
// 4. 用 json.NewDecoder(r.Body).Decode 读请求，json.NewEncoder(w).Encode 写响应。
// 5. 生产环境加 graceful shutdown、限流、超时、认证。

// ── 练习 ──
// 1. 添加 GET /tasks/{id}，用 r.PathValue("id") 取路径参数。
// 2. 添加 DELETE /tasks/{id}，删除指定任务。
// 3. 写一个 authMiddleware，检查 Header "X-Auth-Token"，无效返回 401。
