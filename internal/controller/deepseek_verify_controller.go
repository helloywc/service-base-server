package controller

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
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

type deepseekChatCompletionRequest struct {
	Model    string                   `json:"model"`
	Messages []deepseekChatCompletionMsg `json:"messages"`
	Stream   bool                     `json:"stream"`
}

type deepseekChatCompletionMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepseekVerify 仅 POST
// 1. 从 ai_agent_text 取 key=? 的 content
// 2. 从 bilibili_video 取 status=1、priority 最高的一条的 context
// 3. 用 content/context 调用外部 DeepSeek chat/completions
// 4. 将外部接口返回（状态码+body）透传给客户端
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

	// 与 /api/deepseek 一致：仅本接口放宽写超时为 10 分钟，其他接口仍为全局 15s
	if rc := http.NewResponseController(w); rc != nil {
		_ = rc.SetWriteDeadline(time.Now().Add(deepseekTimeout))
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

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		http.Error(w, "DEEPSEEK_API_KEY is not set", http.StatusServiceUnavailable)
		return
	}

	payload := deepseekChatCompletionRequest{
		Model: "deepseek-chat",
		Messages: []deepseekChatCompletionMsg{
			{Role: "system", Content: content},
			{Role: "user", Content: ctx},
		},
		Stream: false,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "failed to marshal request", http.StatusInternalServerError)
		return
	}

	ctxReq, cancel := context.WithTimeout(r.Context(), deepseekTimeout)
	defer cancel()

	outReq, err := http.NewRequestWithContext(ctxReq, http.MethodPost, "https://api.deepseek.com/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		http.Error(w, "failed to create upstream request", http.StatusInternalServerError)
		return
	}
	outReq.Header.Set("Authorization", "Bearer "+apiKey)
	outReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: deepseekTimeout}
	resp, err := client.Do(outReq)
	if err != nil {
		log.Printf("deepseek-verify: upstream request failed: %v", err)
		http.Error(w, "upstream request failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("deepseek-verify: upstream read failed: %v", err)
		http.Error(w, "upstream read failed", http.StatusBadGateway)
		return
	}

	// 透传外部接口返回
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBody)
}
