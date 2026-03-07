package server

import (
	"net/http"
	"time"

	"code-server/internal/controller"
	"code-server/internal/handler"
	"code-server/internal/service"
)

// New 创建并返回配置好的 HTTP 服务器（兼容 Go 1.21：用路径前缀匹配）
func New(addr string) *http.Server {
	mux := http.NewServeMux()

	// 通用路由
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/", handler.Home)

	// MVC: launchctl bootstrap / bootout（前缀 /api/bootstrap/、/api/bootout/ 匹配 /api/bootstrap/mysql-dev 等）
	launchSvc := service.NewLaunchCtl()
	launchCtrl := controller.NewLaunchController(launchSvc)
	mux.HandleFunc("/api/bootstrap/", launchCtrl.Bootstrap)
	mux.HandleFunc("/api/bootout/", launchCtrl.Bootout)
	mux.HandleFunc("/api/list/", launchCtrl.List)
	mux.HandleFunc("/api/archive/", launchCtrl.Archive)

	return &http.Server{
		Addr:         ":" + addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
}
