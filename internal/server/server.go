package server

import (
	"log"
	"net/http"
	"time"

	"code-server/internal/controller"
	"code-server/internal/db"
	"code-server/internal/handler"
	"code-server/internal/meili"
	"code-server/internal/service"
)

// New 创建并返回配置好的 HTTP 服务器（兼容 Go 1.21：用路径前缀匹配）
func New(addr string) *http.Server {
	mux := http.NewServeMux()
	launchSvc := service.NewLaunchCtl()
	launchCtrl := controller.NewLaunchController(launchSvc)
	contentVerifyCtrl := controller.NewContentVerifyController()

	sqlDB, err := db.OpenMySQL(db.PoolConfig{})
	if err != nil {
		log.Printf("server: MySQL open failed (api/deepseek-verify will return 503): %v", err)
		sqlDB = nil
	}
	deepseekVerifyCtrl := controller.NewDeepseekVerifyController(sqlDB)

	// 精确路径先注册，避免被 "/" 或 "/api/archive/" 抢匹配（Go 1.21 下 longest match 仍可能异常）
	mux.HandleFunc("/api/archives/delete", launchCtrl.DeleteArchives)

	// 通用
	mux.HandleFunc("/health", handler.Health)

	// AI 内容验证（Deepseek）
	// 兼容带不带结尾斜杠的两种访问方式：
	// - /api/deepseek
	// - /api/deepseek/
	mux.HandleFunc("/api/deepseek", contentVerifyCtrl.ContentVerify)
	mux.HandleFunc("/api/deepseek/", contentVerifyCtrl.ContentVerify)

	mux.HandleFunc("/api/deepseek-verify", deepseekVerifyCtrl.DeepseekVerify)
	mux.HandleFunc("/api/deepseek-verify/", deepseekVerifyCtrl.DeepseekVerify)

	// API 前缀路由
	mux.HandleFunc("/api/bootstrap/", launchCtrl.Bootstrap)
	mux.HandleFunc("/api/bootout/", launchCtrl.Bootout)
	mux.HandleFunc("/api/list/", launchCtrl.List)
	mux.HandleFunc("/api/archive/", launchCtrl.Archive)
	mux.HandleFunc("/api/extract/", launchCtrl.Extract)

	// Meilisearch 索引与文档增删改查（环境变量 MEILISEARCH_HOST、MEILISEARCH_API_KEY）
	meiliClient := meili.NewClient()
	meiliCtrl := controller.NewMeiliController(meiliClient)
	mux.HandleFunc("/api/meili/", meiliCtrl.MeiliDispatch)

	// 兜底放最后
	mux.HandleFunc("/", handler.Home)

	return &http.Server{
		Addr:         ":" + addr,
		Handler:      corsHandler(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
}
