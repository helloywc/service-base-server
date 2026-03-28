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
	dbStatusCountCtrl := controller.NewDbStatusCountController(sqlDB)

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

	// /api/deepseek-verify 扩展：后台持续校验任务
	mux.HandleFunc("/api/deepseek-verify/start", deepseekVerifyCtrl.DeepseekVerifyStart)
	mux.HandleFunc("/api/deepseek-verify/stop", deepseekVerifyCtrl.DeepseekVerifyStop)
	mux.HandleFunc("/api/deepseek-verify/status", deepseekVerifyCtrl.DeepseekVerifyStatus)
	mux.HandleFunc("/api/deepseek-verify/last-failure", deepseekVerifyCtrl.DeepseekVerifyLastFailure)

	// MySQL：按表与 status 统计行数（表名白名单）
	mux.HandleFunc("/api/db/status-count", dbStatusCountCtrl.StatusCount)

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
	meiliCtrl := controller.NewMeiliController(meiliClient, sqlDB)
	mux.HandleFunc("/api/meili/", meiliCtrl.MeiliDispatch)

	// Meilisearch 同步（使用 BASE_DB_MEILISEARCH_*，见 .env / .env.development 等）
	mux.HandleFunc("/api/meilisearch/start", meiliCtrl.MeilisearchStart)
	// Meilisearch 搜索代理（支持服务端拼接多 id filter）
	mux.HandleFunc("/api/meilisearch/search", meiliCtrl.MeilisearchSearch)

	// 兜底放最后
	mux.HandleFunc("/", handler.Home)

	return &http.Server{
		Addr:         ":" + addr,
		Handler:      corsHandler(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
}
