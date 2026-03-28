package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"code-server/internal/meili"
	"code-server/internal/view"
)

// MeiliController Meilisearch 索引与文档的增删改查
type MeiliController struct {
	client *meili.Client

	// 用于 /api/meilisearch/start：从 MySQL 拉数据再写入 Meilisearch
	sqlDB *sql.DB
}

// NewMeiliController 创建控制器
func NewMeiliController(client *meili.Client, sqlDB *sql.DB) *MeiliController {
	return &MeiliController{client: client, sqlDB: sqlDB}
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

func buildInPlaceholdersUint64(ids []uint64) (string, []any) {
	if len(ids) == 0 {
		return "", nil
	}
	ph := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		ph[i] = "?"
		args[i] = id
	}
	return strings.Join(ph, ","), args
}

// loadMediaListEnrichment 批量拉取 media_list 名称，及 rel_media_* 关联得到的逗号分隔类型名、标签名
func (c *MeiliController) loadMediaListEnrichment(mediaIDs []uint64) (names, categoryNames, tagNames map[uint64]string, err error) {
	names = make(map[uint64]string)
	categoryNames = make(map[uint64]string)
	tagNames = make(map[uint64]string)
	if len(mediaIDs) == 0 {
		return names, categoryNames, tagNames, nil
	}
	placeholders, args := buildInPlaceholdersUint64(mediaIDs)

	{
		q := fmt.Sprintf("SELECT id, COALESCE(media_name,'') FROM media_list WHERE is_deleted = 0 AND id IN (%s)", placeholders)
		rows, qErr := c.sqlDB.Query(q, args...)
		if qErr != nil {
			return nil, nil, nil, qErr
		}
		defer rows.Close()
		for rows.Next() {
			var id uint64
			var name string
			if scanErr := rows.Scan(&id, &name); scanErr != nil {
				return nil, nil, nil, scanErr
			}
			names[id] = name
		}
		if err := rows.Err(); err != nil {
			return nil, nil, nil, err
		}
	}

	{
		q := fmt.Sprintf(`
SELECT r.media_id, GROUP_CONCAT(DISTINCT mc.category_name ORDER BY mc.id SEPARATOR ',')
FROM rel_media_category r
INNER JOIN media_category mc ON mc.id = r.category_id AND mc.is_deleted = 0
WHERE r.is_deleted = 0 AND r.media_id IN (%s)
GROUP BY r.media_id`, placeholders)
		rows, qErr := c.sqlDB.Query(q, args...)
		if qErr != nil {
			return nil, nil, nil, qErr
		}
		defer rows.Close()
		for rows.Next() {
			var mid uint64
			var list sql.NullString
			if scanErr := rows.Scan(&mid, &list); scanErr != nil {
				return nil, nil, nil, scanErr
			}
			if list.Valid {
				categoryNames[mid] = list.String
			}
		}
		if err := rows.Err(); err != nil {
			return nil, nil, nil, err
		}
	}

	{
		q := fmt.Sprintf(`
SELECT r.media_id, GROUP_CONCAT(DISTINCT mt.tag_name ORDER BY mt.id SEPARATOR ',')
FROM rel_media_tag r
INNER JOIN media_tag mt ON mt.id = r.tag_id AND mt.is_deleted = 0
WHERE r.is_deleted = 0 AND r.media_id IN (%s)
GROUP BY r.media_id`, placeholders)
		rows, qErr := c.sqlDB.Query(q, args...)
		if qErr != nil {
			return nil, nil, nil, qErr
		}
		defer rows.Close()
		for rows.Next() {
			var mid uint64
			var list sql.NullString
			if scanErr := rows.Scan(&mid, &list); scanErr != nil {
				return nil, nil, nil, scanErr
			}
			if list.Valid {
				tagNames[mid] = list.String
			}
		}
		if err := rows.Err(); err != nil {
			return nil, nil, nil, err
		}
	}

	return names, categoryNames, tagNames, nil
}

type meiliSearchStartRequest struct {
	Tabel string `json:"tabel"` // 按你提供的字段名（tabel）
	Table string `json:"table"` // 兼容 table
}

type meiliSearchProxyRequest struct {
	Q       string   `json:"q"`
	Limit   int      `json:"limit"`
	Offset  int      `json:"offset"`
	Index   string   `json:"index,omitempty"`
	IDs     []string `json:"ids,omitempty"`
	IDsCSV  string   `json:"ids_csv,omitempty"`
	// 兼容旧前端：仍然允许传 filter，如 filter = "id = \"a, b, c\""
	Filter  string   `json:"filter,omitempty"`
}

type meiliSearchStartResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (c *MeiliController) writeStartJSON(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(meiliSearchStartResponse{Code: code, Message: message, Data: data})
}

// MeilisearchStart 同步 MySQL 数据到 Meilisearch（仅 POST）
// POST /api/meilisearch/start
// body: {"tabel":"bilibili_video"}
func (c *MeiliController) MeilisearchStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if c.sqlDB == nil {
		c.writeStartJSON(w, 500, "database not available", nil)
		return
	}

	var req meiliSearchStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.writeStartJSON(w, 500, "invalid json body", nil)
		return
	}
	table := strings.TrimSpace(req.Table)
	if table == "" {
		table = strings.TrimSpace(req.Tabel)
	}
	if table == "" {
		c.writeStartJSON(w, 500, "table is required", nil)
		return
	}

	meiliClient := meili.NewClientFromBaseEnv()
	indexUID := strings.TrimSpace(os.Getenv("BASE_DB_MEILISEARCH_INDEX"))
	if indexUID == "" {
		indexUID = "media"
	}

	// 确保索引存在（primaryKey = id）
	_, status, err := meiliClient.IndexGet(indexUID)
	if err != nil && status == http.StatusNotFound {
		_, _, cErr := meiliClient.IndexCreate(indexUID, "id")
		if cErr != nil {
			c.writeStartJSON(w, 500, "meilisearch index create failed: "+cErr.Error(), nil)
			return
		}
	} else if err != nil {
		c.writeStartJSON(w, 500, "meilisearch index get failed: "+err.Error(), nil)
		return
	}

	// 让 `id` 支持过滤：否则使用 filter="id = ..." 会报 invalid_search_filter
	// 说明：primaryKey 不一定自动等价为 filterableAttributes。
	if _, updStatus, updErr := meiliClient.IndexUpdate(indexUID, map[string]any{
		"filterableAttributes": []string{"id"},
	}); updErr != nil {
		c.writeStartJSON(w, 500, "meilisearch index update failed: "+updErr.Error(), map[string]any{
			"status": updStatus,
			"index":  indexUID,
		})
		return
	}

	const batchSize = 50
	switch table {
	case "bilibili_video":
		// 仅同步 status=2 的记录
		rows, qErr := c.sqlDB.Query(
			"SELECT id, COALESCE(video_id,''), COALESCE(title,''), COALESCE(context,''), COALESCE(summary,''), COALESCE(keywords,''), COALESCE(remark,''), "+
				"COALESCE(media_id,''), COALESCE(categories,''), COALESCE(tags,'') "+
				"FROM bilibili_video WHERE is_deleted = 0 AND status = 2 "+
				"ORDER BY id DESC LIMIT ?",
			batchSize,
		)
		if qErr != nil {
			c.writeStartJSON(w, 500, "mysql query failed: "+qErr.Error(), nil)
			return
		}
		defer rows.Close()

		type bvRow struct {
			id         int64
			videoID    string
			title      string
			context    string
			summary    string
			keywords   string
			remark     string
			mediaIDStr string
			categories string
			tags       string
		}

		var records []bvRow
		for rows.Next() {
			var rr bvRow
			if scanErr := rows.Scan(&rr.id, &rr.videoID, &rr.title, &rr.context, &rr.summary, &rr.keywords, &rr.remark, &rr.mediaIDStr, &rr.categories, &rr.tags); scanErr != nil {
				c.writeStartJSON(w, 500, "mysql scan failed: "+scanErr.Error(), nil)
				return
			}
			records = append(records, rr)
		}
		if err := rows.Err(); err != nil {
			c.writeStartJSON(w, 500, "mysql rows error: "+err.Error(), nil)
			return
		}

		seenMedia := make(map[uint64]struct{})
		var mediaIDs []uint64
		for _, rr := range records {
			mid, perr := strconv.ParseUint(strings.TrimSpace(rr.mediaIDStr), 10, 64)
			if perr != nil || mid == 0 {
				continue
			}
			if _, ok := seenMedia[mid]; ok {
				continue
			}
			seenMedia[mid] = struct{}{}
			mediaIDs = append(mediaIDs, mid)
		}

		mediaNames, catNamesByMedia, tagNamesByMedia, enrichErr := c.loadMediaListEnrichment(mediaIDs)
		if enrichErr != nil {
			c.writeStartJSON(w, 500, "mysql media enrichment failed: "+enrichErr.Error(), nil)
			return
		}

		var docs []map[string]interface{}
		var mysqlIDs []int64
		for _, rr := range records {
			docID := fmt.Sprintf("bilibili_video_%d", rr.id)
			mysqlIDs = append(mysqlIDs, rr.id)

			var mediaIDOut, mediaNameOut, catNamesOut, tagNamesOut string
			if mid, perr := strconv.ParseUint(strings.TrimSpace(rr.mediaIDStr), 10, 64); perr == nil && mid > 0 {
				mediaIDOut = fmt.Sprintf("%d", mid)
				mediaNameOut = mediaNames[mid]
				catNamesOut = catNamesByMedia[mid]
				tagNamesOut = tagNamesByMedia[mid]
			}

			docs = append(docs, map[string]interface{}{
				"id":               docID,
				"key":              rr.videoID,
				"title":            rr.title,
				"content":          rr.context,
				"summary":          rr.summary,
				"keywords":         rr.keywords,
				"remark":           rr.remark,
				"status":           1,
				"categories_ids":   rr.categories,
				"tags_ids":         rr.tags,
				"categories_names": catNamesOut,
				"tags_names":       tagNamesOut,
				"media_id":         mediaIDOut,
				"media_name":       mediaNameOut,
			})
		}

		if len(docs) == 0 {
			c.writeStartJSON(w, 200, "no records to sync", map[string]any{"synced": 0})
			return
		}

		_, meiliStatus, meiliErr := meiliClient.DocAdd(indexUID, docs)
		if meiliErr != nil {
			// 写入失败：将这批记录标记为失败（status=-3）
			if len(mysqlIDs) > 0 {
				placeholders := make([]string, 0, len(mysqlIDs))
				args := make([]any, 0, len(mysqlIDs))
				for _, id := range mysqlIDs {
					placeholders = append(placeholders, "?")
					args = append(args, id)
				}
				query := fmt.Sprintf("UPDATE bilibili_video SET status = -3 WHERE id IN (%s)", strings.Join(placeholders, ","))
				_, _ = c.sqlDB.Exec(query, args...)
			}
			c.writeStartJSON(w, 500, fmt.Sprintf("meilisearch doc add failed (status=%d): %v", meiliStatus, meiliErr), nil)
			return
		}

		// 写入成功：标记为已同步（status=3）
		if len(mysqlIDs) > 0 {
			placeholders := make([]string, 0, len(mysqlIDs))
			args := make([]any, 0, len(mysqlIDs))
			for _, id := range mysqlIDs {
				placeholders = append(placeholders, "?")
				args = append(args, id)
			}
			query := fmt.Sprintf("UPDATE bilibili_video SET status = 3 WHERE id IN (%s)", strings.Join(placeholders, ","))
			_, _ = c.sqlDB.Exec(query, args...)
		}

		c.writeStartJSON(w, 200, "ok", map[string]any{
			"synced":   len(docs),
			"index":    indexUID,
			"table":    table,
			"batch":    batchSize,
			"doc_added": "PUT /indexes/:uid/documents",
		})
		return
	default:
		c.writeStartJSON(w, 500, "unsupported table: "+table, map[string]any{"supported": []string{"bilibili_video"}})
		return
	}
}

// MeilisearchSearch 代理搜索并支持服务端拼接多 id filter。
//
// 路由：POST /api/meilisearch/search
// body 示例：
// {
//   "q":"", "limit":20, "offset":0,
//   "ids":["bilibili_video_721","bilibili_video_713"]
// }
func (c *MeiliController) MeilisearchSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.fail(w, http.StatusMethodNotAllowed, "method not allowed, use POST", nil)
		return
	}
	var req meiliSearchProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.fail(w, http.StatusBadRequest, "invalid json body", nil)
		return
	}

	indexUID := strings.TrimSpace(req.Index)
	if indexUID == "" {
		indexUID = strings.TrimSpace(os.Getenv("BASE_DB_MEILISEARCH_INDEX"))
	}
	if indexUID == "" {
		indexUID = "media"
	}

	ids := make([]string, 0, len(req.IDs))
	for _, id := range req.IDs {
		id = strings.TrimSpace(id)
		if id != "" {
			ids = append(ids, id)
		}
	}
	if req.IDsCSV != "" {
		for _, id := range strings.Split(req.IDsCSV, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				ids = append(ids, id)
			}
		}
	}

	searchBody := map[string]any{
		"q":      req.Q,
		"limit":  req.Limit,
		"offset": req.Offset,
	}

	if len(ids) > 0 {
		// 构造 filter：
		// id = "a" OR id = "b" OR ...
		parts := make([]string, 0, len(ids))
		for _, id := range ids {
			// Meilisearch filter 字符串需要转义引号与反斜杠
			escaped := strings.ReplaceAll(id, `\`, `\\`)
			escaped = strings.ReplaceAll(escaped, `"`, `\"`)
			parts = append(parts, fmt.Sprintf("id = \"%s\"", escaped))
		}
		searchBody["filter"] = strings.Join(parts, " OR ")
	} else if strings.TrimSpace(req.Filter) != "" {
		// 兼容：允许传入 id = "a, b, c" 这种逗号拼接字符串
		// 若匹配成功且逗号数量>0，就拆分为 OR 条件。
		filter := strings.TrimSpace(req.Filter)
		// 捕获 id = " ... " 内部内容
		re := regexp.MustCompile(`(?i)id\s*=\s*"(.*?)"`)
		if m := re.FindStringSubmatch(filter); len(m) == 2 {
			rawInner := m[1]
			// inner 内部按逗号拆分，并 trim 空格
			tokens := make([]string, 0, 1)
			for _, tok := range strings.Split(rawInner, ",") {
				tok = strings.TrimSpace(tok)
				if tok != "" {
					tokens = append(tokens, tok)
				}
			}
			if len(tokens) > 1 {
				parts := make([]string, 0, len(tokens))
				for _, id := range tokens {
					escaped := strings.ReplaceAll(id, `\`, `\\`)
					escaped = strings.ReplaceAll(escaped, `"`, `\"`)
					parts = append(parts, fmt.Sprintf("id = \"%s\"", escaped))
				}
				searchBody["filter"] = strings.Join(parts, " OR ")
			} else {
				// 只有一个 token，保留原 filter（保证单个 id 兼容）
				searchBody["filter"] = filter
			}
		} else {
			// 不符合预期格式：直接透传 filter
			searchBody["filter"] = filter
		}
	}

	body, status, err := c.client.Search(indexUID, searchBody)
	if err != nil {
		c.fail(w, status, err.Error(), body)
		return
	}
	c.writeRaw(w, status, body)
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
