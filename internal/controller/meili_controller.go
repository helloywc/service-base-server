package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"code-server/internal/meili"
	"code-server/internal/view"
)

// MeiliController Meilisearch 索引与文档的增删改查
type MeiliController struct {
	client *meili.Client
}

// NewMeiliController 创建控制器
func NewMeiliController(client *meili.Client) *MeiliController {
	return &MeiliController{client: client}
}

func (c *MeiliController) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (c *MeiliController) writeRaw(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
}

func (c *MeiliController) fail(w http.ResponseWriter, code int, msg string, body []byte) {
	resp := view.MeiliErrorResponse{Code: code, Message: msg}
	if len(body) > 0 {
		resp.Body = string(body)
	}
	c.writeJSON(w, code, resp)
}

// 路径解析：/api/meili/indexes/ -> 无后续；/api/meili/indexes/movies -> uid=movies
func meiliIndexUID(path, prefix string) (uid string, rest string) {
	s := strings.TrimPrefix(path, prefix)
	s = strings.Trim(s, "/")
	parts := strings.SplitN(s, "/", 2)
	uid = parts[0]
	if len(parts) > 1 {
		rest = parts[1]
	}
	return uid, rest
}

// IndexList  GET /api/meili/indexes
func (c *MeiliController) IndexList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.fail(w, http.StatusMethodNotAllowed, "method not allowed, use GET", nil)
		return
	}
	body, status, err := c.client.IndexList()
	if err != nil {
		c.fail(w, status, err.Error(), body)
		return
	}
	c.writeRaw(w, status, body)
}

// IndexCreate  POST /api/meili/indexes  body: {"uid":"movies","primaryKey":"id"}
func (c *MeiliController) IndexCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.fail(w, http.StatusMethodNotAllowed, "method not allowed, use POST", nil)
		return
	}
	var req view.IndexCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.fail(w, http.StatusBadRequest, "invalid body: need uid", nil)
		return
	}
	if req.UID == "" {
		c.fail(w, http.StatusBadRequest, "uid required", nil)
		return
	}
	body, status, err := c.client.IndexCreate(req.UID, req.PrimaryKey)
	if err != nil {
		c.fail(w, status, err.Error(), body)
		return
	}
	c.writeRaw(w, status, body)
}

// IndexGet  GET /api/meili/indexes/:uid
func (c *MeiliController) IndexGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.fail(w, http.StatusMethodNotAllowed, "method not allowed, use GET", nil)
		return
	}
	uid, rest := meiliIndexUID(r.URL.Path, "/api/meili/indexes/")
	if uid == "" {
		c.fail(w, http.StatusBadRequest, "index uid required", nil)
		return
	}
	if rest != "" {
		c.fail(w, http.StatusBadRequest, "path should be /api/meili/indexes/:uid", nil)
		return
	}
	body, status, err := c.client.IndexGet(uid)
	if err != nil {
		c.fail(w, status, err.Error(), body)
		return
	}
	c.writeRaw(w, status, body)
}

// IndexUpdate  PUT /api/meili/indexes/:uid  body: {"primaryKey":"id"}
func (c *MeiliController) IndexUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		c.fail(w, http.StatusMethodNotAllowed, "method not allowed, use PUT", nil)
		return
	}
	uid, rest := meiliIndexUID(r.URL.Path, "/api/meili/indexes/")
	if uid == "" || rest != "" {
		c.fail(w, http.StatusBadRequest, "path should be /api/meili/indexes/:uid", nil)
		return
	}
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.fail(w, http.StatusBadRequest, "invalid JSON body", nil)
		return
	}
	body, status, err := c.client.IndexUpdate(uid, req)
	if err != nil {
		c.fail(w, status, err.Error(), body)
		return
	}
	c.writeRaw(w, status, body)
}

// IndexDelete  DELETE /api/meili/indexes/:uid
func (c *MeiliController) IndexDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		c.fail(w, http.StatusMethodNotAllowed, "method not allowed, use DELETE", nil)
		return
	}
	uid, rest := meiliIndexUID(r.URL.Path, "/api/meili/indexes/")
	if uid == "" || rest != "" {
		c.fail(w, http.StatusBadRequest, "path should be /api/meili/indexes/:uid", nil)
		return
	}
	body, status, err := c.client.IndexDelete(uid)
	if err != nil {
		c.fail(w, status, err.Error(), body)
		return
	}
	c.writeRaw(w, status, body)
}

// DocAdd  PUT /api/meili/indexes/:uid/documents  body: [{...}, ...]
func (c *MeiliController) DocAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		c.fail(w, http.StatusMethodNotAllowed, "method not allowed, use PUT", nil)
		return
	}
	uid, rest := meiliIndexUID(r.URL.Path, "/api/meili/indexes/")
	if uid == "" || rest != "documents" {
		c.fail(w, http.StatusBadRequest, "path should be /api/meili/indexes/:uid/documents", nil)
		return
	}
	var docs []map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&docs); err != nil {
		c.fail(w, http.StatusBadRequest, "invalid JSON body: expect array of objects", nil)
		return
	}
	body, status, err := c.client.DocAdd(uid, docs)
	if err != nil {
		c.fail(w, status, err.Error(), body)
		return
	}
	c.writeRaw(w, status, body)
}

// DocList  GET /api/meili/indexes/:uid/documents?limit=20&offset=0
func (c *MeiliController) DocList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.fail(w, http.StatusMethodNotAllowed, "method not allowed, use GET", nil)
		return
	}
	uid, rest := meiliIndexUID(r.URL.Path, "/api/meili/indexes/")
	if uid == "" || rest != "documents" {
		c.fail(w, http.StatusBadRequest, "path should be /api/meili/indexes/:uid/documents", nil)
		return
	}
	limit, offset := 0, 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, _ := parseInt(v); n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, _ := parseInt(v); n >= 0 {
			offset = n
		}
	}
	body, status, err := c.client.DocList(uid, limit, offset)
	if err != nil {
		c.fail(w, status, err.Error(), body)
		return
	}
	c.writeRaw(w, status, body)
}

func parseInt(s string) (int, bool) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	return n, true
}

// MeiliDispatch 统一入口：根据路径和方法分发到对应 handler，注册在 /api/meili/
func (c *MeiliController) MeiliDispatch(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	if path == "/api/meili/indexes" {
		if r.Method == http.MethodGet {
			c.IndexList(w, r)
			return
		}
		if r.Method == http.MethodPost {
			c.IndexCreate(w, r)
			return
		}
		c.fail(w, http.StatusMethodNotAllowed, "use GET or POST", nil)
		return
	}
	if !strings.HasPrefix(path, "/api/meili/indexes/") {
		c.fail(w, http.StatusNotFound, "not found", nil)
		return
	}
	uid, rest := meiliIndexUID(r.URL.Path, "/api/meili/indexes/")
	if uid == "" {
		c.fail(w, http.StatusBadRequest, "index uid required", nil)
		return
	}
	switch rest {
	case "":
		switch r.Method {
		case http.MethodGet:
			c.IndexGet(w, r)
		case http.MethodPut:
			c.IndexUpdate(w, r)
		case http.MethodDelete:
			c.IndexDelete(w, r)
		default:
			c.fail(w, http.StatusMethodNotAllowed, "use GET, PUT or DELETE", nil)
		}
	case "documents":
		switch r.Method {
		case http.MethodGet:
			c.DocList(w, r)
		case http.MethodPut:
			c.DocAdd(w, r)
		case http.MethodDelete:
			c.DocDeleteAll(w, r)
		default:
			c.fail(w, http.StatusMethodNotAllowed, "use GET, PUT or DELETE", nil)
		}
	case "documents/delete-batch":
		c.DocDeleteBatch(w, r)
	default:
		if strings.HasPrefix(rest, "documents/") {
			c.docOne(w, r, r.URL.Path)
			return
		}
		c.fail(w, http.StatusBadRequest, "unknown path", nil)
	}
}

func (c *MeiliController) docOne(w http.ResponseWriter, r *http.Request, path string) {
	_, rest := meiliIndexUID(path, "/api/meili/indexes/")
	if !strings.HasPrefix(rest, "documents/") {
		c.fail(w, http.StatusBadRequest, "path should be .../documents/:id", nil)
		return
	}
	docID := strings.TrimPrefix(rest, "documents/")
	if docID == "" {
		c.fail(w, http.StatusBadRequest, "document id required", nil)
		return
	}
	if r.Method == http.MethodGet {
		c.DocGet(w, r)
		return
	}
	if r.Method == http.MethodDelete {
		c.DocDeleteOne(w, r)
		return
	}
	c.fail(w, http.StatusMethodNotAllowed, "use GET or DELETE", nil)
}

// DocGet  GET /api/meili/indexes/:uid/documents/:documentId
func (c *MeiliController) DocGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.fail(w, http.StatusMethodNotAllowed, "method not allowed, use GET", nil)
		return
	}
	uid, rest := meiliIndexUID(r.URL.Path, "/api/meili/indexes/")
	if uid == "" || rest == "" {
		c.fail(w, http.StatusBadRequest, "path should be /api/meili/indexes/:uid/documents/:documentId", nil)
		return
	}
	if !strings.HasPrefix(rest, "documents/") {
		c.fail(w, http.StatusBadRequest, "path should be /api/meili/indexes/:uid/documents/:documentId", nil)
		return
	}
	docID := strings.TrimPrefix(rest, "documents/")
	if docID == "" {
		c.fail(w, http.StatusBadRequest, "document id required", nil)
		return
	}
	body, status, err := c.client.DocGet(uid, docID)
	if err != nil {
		c.fail(w, status, err.Error(), body)
		return
	}
	c.writeRaw(w, status, body)
}

// DocDeleteOne  DELETE /api/meili/indexes/:uid/documents/:documentId
func (c *MeiliController) DocDeleteOne(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		c.fail(w, http.StatusMethodNotAllowed, "method not allowed, use DELETE", nil)
		return
	}
	uid, rest := meiliIndexUID(r.URL.Path, "/api/meili/indexes/")
	if uid == "" || rest == "" {
		c.fail(w, http.StatusBadRequest, "path should be /api/meili/indexes/:uid/documents/:documentId", nil)
		return
	}
	if !strings.HasPrefix(rest, "documents/") {
		c.fail(w, http.StatusBadRequest, "path should be /api/meili/indexes/:uid/documents/:documentId", nil)
		return
	}
	docID := strings.TrimPrefix(rest, "documents/")
	if docID == "" {
		c.fail(w, http.StatusBadRequest, "document id required", nil)
		return
	}
	body, status, err := c.client.DocDeleteOne(uid, docID)
	if err != nil {
		c.fail(w, status, err.Error(), body)
		return
	}
	c.writeRaw(w, status, body)
}

// DocDeleteBatch  POST /api/meili/indexes/:uid/documents/delete-batch  body: {"ids":["1","2"]}
func (c *MeiliController) DocDeleteBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.fail(w, http.StatusMethodNotAllowed, "method not allowed, use POST", nil)
		return
	}
	uid, rest := meiliIndexUID(r.URL.Path, "/api/meili/indexes/")
	if uid == "" || rest != "documents/delete-batch" {
		c.fail(w, http.StatusBadRequest, "path should be /api/meili/indexes/:uid/documents/delete-batch", nil)
		return
	}
	var req struct {
		Ids []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.fail(w, http.StatusBadRequest, "invalid body: need { ids: [] }", nil)
		return
	}
	body, status, err := c.client.DocDeleteBatch(uid, req.Ids)
	if err != nil {
		c.fail(w, status, err.Error(), body)
		return
	}
	c.writeRaw(w, status, body)
}

// DocDeleteAll  DELETE /api/meili/indexes/:uid/documents
func (c *MeiliController) DocDeleteAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		c.fail(w, http.StatusMethodNotAllowed, "method not allowed, use DELETE", nil)
		return
	}
	uid, rest := meiliIndexUID(r.URL.Path, "/api/meili/indexes/")
	if uid == "" || rest != "documents" {
		c.fail(w, http.StatusBadRequest, "path should be /api/meili/indexes/:uid/documents", nil)
		return
	}
	body, status, err := c.client.DocDeleteAll(uid)
	if err != nil {
		c.fail(w, status, err.Error(), body)
		return
	}
	c.writeRaw(w, status, body)
}
