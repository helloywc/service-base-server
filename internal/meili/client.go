package meili

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"net/http"
	"os"
	"strings"
)

const (
	envHost = "MEILISEARCH_HOST"
	envKey  = "MEILISEARCH_API_KEY"
)

// Client 调用 Meilisearch REST API
type Client struct {
	host   string
	apiKey string
	http   *http.Client
}

// NewClient 从环境变量读取 MEILISEARCH_HOST（默认 http://localhost:7700）、MEILISEARCH_API_KEY（默认 123456）
func NewClient() *Client {
	host := os.Getenv(envHost)
	if host == "" {
		host = "http://localhost:7700"
	}
	key := os.Getenv(envKey)
	if key == "" {
		key = "123456"
	}
	return &Client{
		host:   host,
		apiKey: key,
		http:   &http.Client{},
	}
}

// NewClientWithHostKey 创建自定义连接（便于从其他 env 变量读取）
func NewClientWithHostKey(host, key string) *Client {
	if host == "" {
		host = "http://localhost:7700"
	}
	if key == "" {
		key = "123456"
	}
	return &Client{
		host:   host,
		apiKey: key,
		http:   &http.Client{},
	}
}

// NewClientFromBaseEnv 从 .env.dev 的 BASE_DB_MEILISEARCH_* 变量读取连接配置
func NewClientFromBaseEnv() *Client {
	baseURL := os.Getenv("BASE_DB_MEILISEARCH_URL")
	basePort := os.Getenv("BASE_DB_MEILISEARCH_PORT")
	baseKey := os.Getenv("BASE_DB_MEILISEARCH_MASTER_KEY")
	if baseURL == "" {
		// 回退到原逻辑（MEILISEARCH_HOST/MEILISEARCH_API_KEY）
		return NewClient()
	}

	host := strings.TrimRight(baseURL, "/")
	if basePort != "" {
		if u, err := url.Parse(host); err == nil && u.Scheme != "" {
			u.Host = u.Hostname() + ":" + basePort
			host = u.String()
		} else {
			host = host + ":" + basePort
		}
	}
	return NewClientWithHostKey(host, baseKey)
}

func (c *Client) do(method, path string, body any) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		reqBody = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.host+path, reqBody)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if resp.StatusCode >= 400 {
		return out, resp.StatusCode, fmt.Errorf("meilisearch: %s", string(out))
	}
	return out, resp.StatusCode, nil
}

// Get  GET 请求
func (c *Client) Get(path string) ([]byte, int, error) {
	return c.do(http.MethodGet, path, nil)
}

// Post POST 请求
func (c *Client) Post(path string, body any) ([]byte, int, error) {
	return c.do(http.MethodPost, path, body)
}

// Put PUT 请求
func (c *Client) Put(path string, body any) ([]byte, int, error) {
	return c.do(http.MethodPut, path, body)
}

// Delete DELETE 请求
func (c *Client) Delete(path string) ([]byte, int, error) {
	return c.do(http.MethodDelete, path, nil)
}
