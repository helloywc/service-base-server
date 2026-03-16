package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"
)

type ContentVerifyController struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

func NewContentVerifyController() *ContentVerifyController {
	baseURL := os.Getenv("DEEPSEEK_ADDRESS")
	apiKey := os.Getenv("DEEPSEEK_API_KEY")

	return &ContentVerifyController{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

type contentVerifyRequest struct {
	Prompt  string `json:"prompt"`
	Content string `json:"content"`
}

// ContentVerify 透传前端 prompt / content 到 Deepseek /chat/completions
// 请求体：
// {
//   "prompt": "<system prompt>",
//   "content": "<user content>"
// }
func (c *ContentVerifyController) ContentVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if c.baseURL == "" || c.apiKey == "" {
		http.Error(w, "deepseek config not set", http.StatusInternalServerError)
		return
	}

	var reqBody contentVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if reqBody.Prompt == "" || reqBody.Content == "" {
		http.Error(w, "prompt and content are required", http.StatusBadRequest)
		return
	}

	payload := map[string]any{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": reqBody.Prompt},
			{"role": "user", "content": reqBody.Content},
		},
		"response_format": map[string]string{"type": "json_object"},
		"stream":          false,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "failed to marshal payload", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	url := c.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		http.Error(w, "failed to call deepseek", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read deepseek response", http.StatusBadGateway)
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// 简单透传 Deepseek 的错误信息
		http.Error(w, string(respBody), http.StatusBadGateway)
		return
	}

	// 直接把 Deepseek 的 JSON 结果透传给前端
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(respBody); err != nil {
		_ = errors.New("write response failed")
	}
}

