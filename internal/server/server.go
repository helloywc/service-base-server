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
	launchSvc := service.NewLaunchCtl()
	launchCtrl := controller.NewLaunchController(launchSvc)

	// 精确路径先注册，避免被 "/" 或 "/api/archive/" 抢匹配（Go 1.21 下 longest match 仍可能异常）
	mux.HandleFunc("/api/archives/delete", launchCtrl.DeleteArchives)

	// 通用
	mux.HandleFunc("/health", handler.Health)

	// API 前缀路由
	mux.HandleFunc("/api/bootstrap/", launchCtrl.Bootstrap)
	mux.HandleFunc("/api/bootout/", launchCtrl.Bootout)
	mux.HandleFunc("/api/list/", launchCtrl.List)
	mux.HandleFunc("/api/archive/", launchCtrl.Archive)
	mux.HandleFunc("/api/extract/", launchCtrl.Extract)

	// 兜底放最后
	mux.HandleFunc("/", handler.Home)

	return &http.Server{
		Addr:         ":" + addr,
		Handler:      corsHandler(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
}
