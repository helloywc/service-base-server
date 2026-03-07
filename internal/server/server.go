package server

import (
	"net/http"
	"time"

	"code-server/internal/controller"
	"code-server/internal/handler"
	"code-server/internal/service"
)

// New 创建并返回配置好的 HTTP 服务器
func New(addr string) *http.Server {
	mux := http.NewServeMux()

	// 通用路由
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("GET /", handler.Home)

	// MVC: launchctl bootstrap / bootout
	launchSvc := service.NewLaunchCtl()
	launchCtrl := controller.NewLaunchController(launchSvc)
	mux.HandleFunc("POST /api/bootstrap/{name}", launchCtrl.Bootstrap)
	mux.HandleFunc("POST /api/bootout/{name}", launchCtrl.Bootout)

	return &http.Server{
		Addr:         ":" + addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
}
