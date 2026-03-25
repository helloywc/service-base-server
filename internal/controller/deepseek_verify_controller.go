package controller

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// DeepseekVerifyController 处理 /api/deepseek-verify：从 DB 取 content/context 并返回
type DeepseekVerifyController struct {
	sqlDB *sql.DB

	mu             sync.Mutex
	running        bool
	cancel         context.CancelFunc
	key            string
	currentVideoID int64 // 当前正在处理的一条记录 id，停止任务时用于写入失败原因
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
	Running        bool   `json:"running"`
	Key            string `json:"key,omitempty"`
	WasRunning     *bool  `json:"was_running,omitempty"`      // stop：调用前是否在运行
	StoppedVideoID int64  `json:"stopped_video_id,omitempty"` // stop：被中止时正在处理的 bilibili_video.id
}

type deepseekVerifyAPIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func writeDeepseekVerifyJSON(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(deepseekVerifyAPIResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// 模型返回的 message.content 可能是 JSON：{"content":"...", "summary":"..."}
type deepseekMessageContent struct {
	Content  string          `json:"content"`
	Summary  string          `json:"summary"`
	Keywords json.RawMessage `json:"keywords"`
}

// parseMessageContent 解析 choices[0].message.content：若为 JSON 则拆成 content/summary，否则整段作为 content。
func parseMessageContent(raw string) (content, summary, keywords string) {
	content = raw
	if raw == "" {
		return "", "", ""
	}
	var parsed deepseekMessageContent
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return content, "", ""
	}
	if parsed.Content != "" {
		content = parsed.Content
	}
	// summary 表字段为 VARCHAR(255)
	if len(parsed.Summary) > 255 {
		summary = parsed.Summary[:255]
	} else {
		summary = parsed.Summary
	}
	keywords = normalizeKeywords(parsed.Keywords)
	return content, summary, keywords
}

func normalizeKeywords(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return ""
	}

	var asString string
	if err := json.Unmarshal(trimmed, &asString); err == nil {
		return truncateKeywords(asString)
	}

	var asArray []string
	if err := json.Unmarshal(trimmed, &asArray); err == nil {
		if len(asArray) == 0 {
			return ""
		}
		return truncateKeywords(strings.Join(asArray, ","))
	}

	return truncateKeywords(string(trimmed))
}

func truncateKeywords(s string) string {
	// keywords 表字段为 VARCHAR(255)
	if len(s) > 255 {
		return s[:255]
	}
	return s
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

func (d *DeepseekVerifyController) getCurrentVideoID() int64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.currentVideoID
}

func (d *DeepseekVerifyController) setCurrentVideoID(videoID int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.currentVideoID = videoID
}

func (d *DeepseekVerifyController) stopTask() (wasRunning bool, videoID int64) {
	d.mu.Lock()
	cancel := d.cancel
	wasRunning = d.running
	videoID = d.currentVideoID
	d.running = false
	d.key = ""
	d.cancel = nil
	d.currentVideoID = 0
	d.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	return wasRunning, videoID
}

func (d *DeepseekVerifyController) getPromptByKey(key string) (string, error) {
	var prompt string
	err := d.sqlDB.QueryRow(
		"SELECT content FROM ai_agent_text WHERE `key` = ? LIMIT 1",
		key,
	).Scan(&prompt)
	return prompt, err
}

const maxRemarkLen = 512

func truncateReason(s string) string {
	if len(s) <= maxRemarkLen {
		return s
	}
	return s[:maxRemarkLen]
}

func (d *DeepseekVerifyController) setVerifyFailed(videoID int64, reason string) {
	_, _ = d.sqlDB.Exec("UPDATE bilibili_video SET status = -2, remark = ? WHERE id = ?", truncateReason(reason), videoID)
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
	d.setCurrentVideoID(videoID)
	defer d.setCurrentVideoID(0)

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
		d.setVerifyFailed(videoID, "failed to marshal request: "+mErr.Error())
		return true, mErr
	}

	reqCtx, cancel := context.WithTimeout(parentCtx, deepseekTimeout)
	defer cancel()

	outReq, rErr := http.NewRequestWithContext(reqCtx, http.MethodPost, "https://api.deepseek.com/chat/completions", bytes.NewReader(bodyBytes))
	if rErr != nil {
		d.setVerifyFailed(videoID, "failed to create upstream request: "+rErr.Error())
		return true, rErr
	}
	outReq.Header.Set("Authorization", "Bearer "+apiKey)
	outReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: deepseekTimeout}
	resp, doErr := client.Do(outReq)
	if doErr != nil {
		d.setVerifyFailed(videoID, "upstream request failed: "+doErr.Error())
		return true, doErr
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		d.setVerifyFailed(videoID, "upstream read failed: "+readErr.Error())
		return true, readErr
	}

	logDeepseekUpstream("deepseek-verify/start", resp.StatusCode, respBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		d.setVerifyFailed(videoID, "upstream returned non-2xx: "+fmt.Sprintf("%d", resp.StatusCode))
		return true, nil
	}

	var dsResp deepseekChatCompletionResponse
	if err := json.Unmarshal(respBody, &dsResp); err != nil || len(dsResp.Choices) == 0 {
		d.setVerifyFailed(videoID, "invalid deepseek response")
		return true, nil
	}
	rawContent := dsResp.Choices[0].Message.Content
	if rawContent == "" {
		d.setVerifyFailed(videoID, "deepseek response content is empty")
		return true, nil
	}
	content, summary, keywords := parseMessageContent(rawContent)
	if _, err := d.sqlDB.Exec("UPDATE bilibili_video SET context = ?, summary = ?, keywords = ?, status = 2 WHERE id = ?", content, summary, keywords, videoID); err != nil {
		d.setVerifyFailed(videoID, "db update failed: "+err.Error())
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
		writeDeepseekVerifyJSON(w, 500, "database not available", nil)
		return
	}

	var req deepseekVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDeepseekVerifyJSON(w, 500, "invalid json body", nil)
		return
	}
	if req.Key == "" {
		writeDeepseekVerifyJSON(w, 500, "key is required", nil)
		return
	}

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		writeDeepseekVerifyJSON(w, 500, "DEEPSEEK_API_KEY is not set", nil)
		return
	}

	d.mu.Lock()
	if d.running {
		key := d.key
		d.mu.Unlock()
		data := deepseekVerifyTaskStatusResponse{Running: true, Key: key}
		writeDeepseekVerifyJSON(w, 500, fmt.Sprintf("已有任务在运行，当前 key=%s", key), data)
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

	data := deepseekVerifyTaskStatusResponse{Running: true, Key: req.Key}
	writeDeepseekVerifyJSON(w, 200, fmt.Sprintf("后台校验任务已启动，key=%s，运行中 running=true", req.Key), data)
}

// DeepseekVerifyStop：终止后台任务；若当前有正在处理的记录，将其标记为失败并在 remark 写入原因
func (d *DeepseekVerifyController) DeepseekVerifyStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	wasRunning, videoID := d.stopTask()
	if videoID != 0 && d.sqlDB != nil {
		d.setVerifyFailed(videoID, "任务已由用户停止")
	}
	wr := wasRunning
	data := deepseekVerifyTaskStatusResponse{
		Running:        false,
		WasRunning:     &wr,
		StoppedVideoID: videoID,
	}
	var msg string
	switch {
	case wasRunning && videoID != 0:
		msg = fmt.Sprintf("后台任务已停止，此前正在处理 bilibili_video.id=%d，已标记为整理失败", videoID)
	case wasRunning:
		msg = "后台任务已停止（此前无正在处理的单条记录）"
	default:
		msg = "当前没有运行中的后台任务，无需停止"
	}
	writeDeepseekVerifyJSON(w, 200, msg, data)
}

// DeepseekVerifyStatus：查询后台任务状态
func (d *DeepseekVerifyController) DeepseekVerifyStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	running, key := d.getTaskStatus()
	data := deepseekVerifyTaskStatusResponse{Running: running, Key: key}
	var msg string
	if running {
		if key != "" {
			msg = fmt.Sprintf("任务运行中，running=true，key=%s", key)
		} else {
			msg = "任务运行中，running=true"
		}
	} else {
		msg = "任务未运行，running=false"
	}
	writeDeepseekVerifyJSON(w, 200, msg, data)
}

// lastVerifyFailure 上一次整理失败记录（供 /api/deepseek-verify/last-failure 返回）
type lastVerifyFailure struct {
	ID            int64  `json:"id"`
	VideoID       string `json:"video_id"`
	Title         string `json:"title,omitempty"`
	Reason        string `json:"reason"` // 失败原因，来自 remark
	LastModifyTs  int64  `json:"last_modify_ts"`
}

// DeepseekVerifyLastFailure：查询上一次整理失败的记录及原因
func (d *DeepseekVerifyController) DeepseekVerifyLastFailure(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if d.sqlDB == nil {
		writeDeepseekVerifyJSON(w, 500, "database not available", nil)
		return
	}
	var rec lastVerifyFailure
	err := d.sqlDB.QueryRow(
		"SELECT id, video_id, COALESCE(title,''), COALESCE(remark,''), last_modify_ts FROM bilibili_video WHERE status = -2 ORDER BY id DESC LIMIT 1",
	).Scan(&rec.ID, &rec.VideoID, &rec.Title, &rec.Reason, &rec.LastModifyTs)
	if err != nil {
		if err == sql.ErrNoRows {
			writeDeepseekVerifyJSON(w, 200, "ok", nil)
			return
		}
		log.Printf("deepseek-verify/last-failure: query error: %v", err)
		writeDeepseekVerifyJSON(w, 500, "db error", nil)
		return
	}
	writeDeepseekVerifyJSON(w, 200, "ok", rec)
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
		d.setVerifyFailed(videoID, "upstream request failed: "+err.Error())
		http.Error(w, "upstream request failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("deepseek-verify: upstream read failed: %v", err)
		d.setVerifyFailed(videoID, "upstream read failed: "+err.Error())
		http.Error(w, "upstream read failed", http.StatusBadGateway)
		return
	}

	logDeepseekUpstream("deepseek-verify", resp.StatusCode, respBody)

	// 如果上游返回非 2xx，也视为失败：标记 status=-2，但仍透传上游响应
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		d.setVerifyFailed(videoID, "upstream returned non-2xx: "+fmt.Sprintf("%d", resp.StatusCode))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(respBody)
		return
	}

	// 解析上游响应的 choices[0].message.content（支持 JSON {"content":"...","summary":"..."}），回写 context/summary，并标记 status=2
	var dsResp deepseekChatCompletionResponse
	if err := json.Unmarshal(respBody, &dsResp); err != nil || len(dsResp.Choices) == 0 {
		d.setVerifyFailed(videoID, "invalid deepseek response")
		http.Error(w, "invalid deepseek response", http.StatusBadGateway)
		return
	}
	rawContent := dsResp.Choices[0].Message.Content
	if rawContent == "" {
		d.setVerifyFailed(videoID, "deepseek response content is empty")
		http.Error(w, "deepseek response content is empty", http.StatusBadGateway)
		return
	}
	content, summary, keywords := parseMessageContent(rawContent)
	if _, err := d.sqlDB.Exec("UPDATE bilibili_video SET context = ?, summary = ?, keywords = ?, status = 2 WHERE id = ?", content, summary, keywords, videoID); err != nil {
		log.Printf("deepseek-verify: bilibili_video update error: %v", err)
		d.setVerifyFailed(videoID, "db update failed: "+err.Error())
		http.Error(w, "db update failed", http.StatusInternalServerError)
		return
	}

	// 透传外部接口返回
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBody)
}
