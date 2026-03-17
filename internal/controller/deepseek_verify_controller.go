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
	"sync"
	"time"
)

// DeepseekVerifyController 处理 /api/deepseek-verify：从 DB 取 content/context 并返回
type DeepseekVerifyController struct {
	sqlDB *sql.DB

	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc
	key     string
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

type deepseekChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type deepseekVerifyTaskStatusResponse struct {
	Running bool   `json:"running"`
	Key     string `json:"key,omitempty"`
}

func truncateForLog(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	return s[:max] + "...(truncated)"
}

func logDeepseekUpstream(tag string, statusCode int, body []byte) {
	// 避免日志刷屏：最多打印 2KB
	log.Printf("%s: deepseek upstream status=%d body=%s", tag, statusCode, truncateForLog(string(body), 2048))
}

func (d *DeepseekVerifyController) getTaskStatus() (running bool, key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.running, d.key
}

func (d *DeepseekVerifyController) setTask(running bool, key string, cancel context.CancelFunc) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.running = running
	d.key = key
	d.cancel = cancel
}

func (d *DeepseekVerifyController) stopTask() (wasRunning bool) {
	d.mu.Lock()
	cancel := d.cancel
	wasRunning = d.running
	d.running = false
	d.key = ""
	d.cancel = nil
	d.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	return wasRunning
}

func (d *DeepseekVerifyController) getPromptByKey(key string) (string, error) {
	var prompt string
	err := d.sqlDB.QueryRow(
		"SELECT content FROM ai_agent_text WHERE `key` = ? LIMIT 1",
		key,
	).Scan(&prompt)
	return prompt, err
}

// processNextOne: 取一条 bilibili_video(status=1, priority最高) 并尝试调用 DeepSeek。
// 返回 found=false 表示已经没有可处理记录，应停止循环。
func (d *DeepseekVerifyController) processNextOne(parentCtx context.Context, apiKey string, prompt string) (found bool, err error) {
	var (
		videoID int64
		ctx     string
	)
	qErr := d.sqlDB.QueryRow(
		"SELECT id, context FROM bilibili_video WHERE status = 1 ORDER BY priority DESC LIMIT 1",
	).Scan(&videoID, &ctx)
	if qErr != nil {
		if qErr == sql.ErrNoRows {
			return false, nil
		}
		return true, qErr
	}

	payload := deepseekChatCompletionRequest{
		Model: "deepseek-chat",
		Messages: []deepseekChatCompletionMsg{
			{Role: "system", Content: prompt},
			{Role: "user", Content: ctx},
		},
		Stream: false,
	}
	bodyBytes, mErr := json.Marshal(payload)
	if mErr != nil {
		_, _ = d.sqlDB.Exec("UPDATE bilibili_video SET status = -2 WHERE id = ?", videoID)
		return true, mErr
	}

	reqCtx, cancel := context.WithTimeout(parentCtx, deepseekTimeout)
	defer cancel()

	outReq, rErr := http.NewRequestWithContext(reqCtx, http.MethodPost, "https://api.deepseek.com/chat/completions", bytes.NewReader(bodyBytes))
	if rErr != nil {
		_, _ = d.sqlDB.Exec("UPDATE bilibili_video SET status = -2 WHERE id = ?", videoID)
		return true, rErr
	}
	outReq.Header.Set("Authorization", "Bearer "+apiKey)
	outReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: deepseekTimeout}
	resp, doErr := client.Do(outReq)
	if doErr != nil {
		_, _ = d.sqlDB.Exec("UPDATE bilibili_video SET status = -2 WHERE id = ?", videoID)
		return true, doErr
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		_, _ = d.sqlDB.Exec("UPDATE bilibili_video SET status = -2 WHERE id = ?", videoID)
		return true, readErr
	}

	logDeepseekUpstream("deepseek-verify/start", resp.StatusCode, respBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = d.sqlDB.Exec("UPDATE bilibili_video SET status = -2 WHERE id = ?", videoID)
		return true, nil
	}

	var dsResp deepseekChatCompletionResponse
	if err := json.Unmarshal(respBody, &dsResp); err != nil || len(dsResp.Choices) == 0 {
		_, _ = d.sqlDB.Exec("UPDATE bilibili_video SET status = -2 WHERE id = ?", videoID)
		return true, nil
	}
	newCtx := dsResp.Choices[0].Message.Content
	if newCtx == "" {
		_, _ = d.sqlDB.Exec("UPDATE bilibili_video SET status = -2 WHERE id = ?", videoID)
		return true, nil
	}
	if _, err := d.sqlDB.Exec("UPDATE bilibili_video SET context = ?, status = 2 WHERE id = ?", newCtx, videoID); err != nil {
		_, _ = d.sqlDB.Exec("UPDATE bilibili_video SET status = -2 WHERE id = ?", videoID)
		return true, err
	}

	return true, nil
}

// DeepseekVerifyStart：启动后台循环处理，直到没有 status=1 的记录。
func (d *DeepseekVerifyController) DeepseekVerifyStart(w http.ResponseWriter, r *http.Request) {
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

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		http.Error(w, "DEEPSEEK_API_KEY is not set", http.StatusServiceUnavailable)
		return
	}

	d.mu.Lock()
	if d.running {
		key := d.key
		d.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(deepseekVerifyTaskStatusResponse{Running: true, Key: key})
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	d.running = true
	d.key = req.Key
	d.cancel = cancel
	d.mu.Unlock()

	go func(key string) {
		defer func() {
			d.setTask(false, "", nil)
		}()

		prompt, err := d.getPromptByKey(key)
		if err != nil {
			log.Printf("deepseek-verify/start: ai_agent_text query error: %v", err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			found, err := d.processNextOne(ctx, apiKey, prompt)
			if err != nil {
				log.Printf("deepseek-verify/start: process error: %v", err)
			}
			if !found {
				return
			}
		}
	}(req.Key)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(deepseekVerifyTaskStatusResponse{Running: true, Key: req.Key})
}

// DeepseekVerifyStop：终止后台任务
func (d *DeepseekVerifyController) DeepseekVerifyStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = d.stopTask()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(deepseekVerifyTaskStatusResponse{Running: false})
}

// DeepseekVerifyStatus：查询后台任务状态
func (d *DeepseekVerifyController) DeepseekVerifyStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	running, key := d.getTaskStatus()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(deepseekVerifyTaskStatusResponse{Running: running, Key: key})
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
	content, err := d.getPromptByKey(req.Key)
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
	var (
		videoID int64
		ctx     string
	)
	err = d.sqlDB.QueryRow(
		"SELECT id, context FROM bilibili_video WHERE status = 1 ORDER BY priority DESC LIMIT 1",
	).Scan(&videoID, &ctx)
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

	// 单次接口：沿用旧行为（透传上游响应），但 DB 更新逻辑复用 processNextOne
	// 注意：这里已经选出了具体 videoID/context，因此单次接口继续沿用原实现以确保透传语义一致。
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
		_, _ = d.sqlDB.Exec("UPDATE bilibili_video SET status = -2 WHERE id = ?", videoID)
		http.Error(w, "upstream request failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("deepseek-verify: upstream read failed: %v", err)
		_, _ = d.sqlDB.Exec("UPDATE bilibili_video SET status = -2 WHERE id = ?", videoID)
		http.Error(w, "upstream read failed", http.StatusBadGateway)
		return
	}

	logDeepseekUpstream("deepseek-verify", resp.StatusCode, respBody)

	// 如果上游返回非 2xx，也视为失败：标记 status=-2，但仍透传上游响应
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = d.sqlDB.Exec("UPDATE bilibili_video SET status = -2 WHERE id = ?", videoID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(respBody)
		return
	}

	// 解析上游响应的 choices[0].message.content，并回写到 bilibili_video.context，同时标记 status=2
	var dsResp deepseekChatCompletionResponse
	if err := json.Unmarshal(respBody, &dsResp); err != nil || len(dsResp.Choices) == 0 {
		_, _ = d.sqlDB.Exec("UPDATE bilibili_video SET status = -2 WHERE id = ?", videoID)
		http.Error(w, "invalid deepseek response", http.StatusBadGateway)
		return
	}
	newCtx := dsResp.Choices[0].Message.Content
	if newCtx == "" {
		_, _ = d.sqlDB.Exec("UPDATE bilibili_video SET status = -2 WHERE id = ?", videoID)
		http.Error(w, "deepseek response content is empty", http.StatusBadGateway)
		return
	}
	if _, err := d.sqlDB.Exec("UPDATE bilibili_video SET context = ?, status = 2 WHERE id = ?", newCtx, videoID); err != nil {
		log.Printf("deepseek-verify: bilibili_video update error: %v", err)
		_, _ = d.sqlDB.Exec("UPDATE bilibili_video SET status = -2 WHERE id = ?", videoID)
		http.Error(w, "db update failed", http.StatusInternalServerError)
		return
	}

	// 透传外部接口返回
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBody)
}
