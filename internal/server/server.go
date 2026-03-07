package server

import (
	"net/http"
	"time"

	"code-server/internal/handler"
)

// New 创建并返回配置好的 HTTP 服务器
func New(addr string) *http.Server {
	mux := http.NewServeMux()

	// 注册路由
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("GET /", handler.Home)

	return &http.Server{
		Addr:         ":" + addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
}
