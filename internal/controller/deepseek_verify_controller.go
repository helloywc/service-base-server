package controller

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

// DeepseekVerifyController 处理 /api/deepseek-verify：从 DB 取 content/context 并返回
type DeepseekVerifyController struct {
	sqlDB *sql.DB
}

// NewDeepseekVerifyController 依赖 MySQL 连接池，sqlDB 可为 nil（接口返回 503）
func NewDeepseekVerifyController(sqlDB *sql.DB) *DeepseekVerifyController {
	return &DeepseekVerifyController{sqlDB: sqlDB}
}

// 请求体
type deepseekVerifyRequest struct {
	Key   string `json:"key"`   // 如 "text-verify"，查 ai_agent_text 的 key 字段
	Table string `json:"table"` // 如 "bilibili_video"，当前仅用表名约定，实际查 bilibili_video
}

// 响应体：返回的两个值
type deepseekVerifyResponse struct {
	Content string `json:"content"` // 来自 ai_agent_text.content
	Context string `json:"context"` // 来自 bilibili_video.context
}

// DeepseekVerify 仅 POST
// 1. 从 ai_agent_text 取 key=? 的 content
// 2. 从 bilibili_video 取 status=1、priority 最高的一条的 context
// 3. 返回 { "content": "...", "context": "..." }
func (d *DeepseekVerifyController) DeepseekVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if d.sqlDB == nil {
		http.Error(w, "database not available", http.StatusServiceUnavailable)
		return
	}

	var req deepseekVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if req.Key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	// 1. ai_agent_text：key 字段 = req.Key，取 content
	var content string
	err := d.sqlDB.QueryRow(
		"SELECT content FROM ai_agent_text WHERE `key` = ? LIMIT 1",
		req.Key,
	).Scan(&content)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "ai_agent_text: no row for key", http.StatusBadRequest)
			return
		}
		log.Printf("deepseek-verify: ai_agent_text query error: %v", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	// 2. bilibili_video：status=1，按 priority 降序，取第一条的 context
	var ctx string
	err = d.sqlDB.QueryRow(
		"SELECT context FROM bilibili_video WHERE status = 1 ORDER BY priority DESC LIMIT 1",
	).Scan(&ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "bilibili_video: no row with status=1", http.StatusBadRequest)
			return
		}
		log.Printf("deepseek-verify: bilibili_video query error: %v", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(deepseekVerifyResponse{Content: content, Context: ctx})
}
