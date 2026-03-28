package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// 仅允许访问的表名（key 为小写，便于与请求体统一比较；value 为真实 SQL 标识）
var allowedStatusCountTables = map[string]string{
	"bilibili_video": "bilibili_video",
}

// DbStatusCountController 按表名与 status 统计行数（COUNT）
type DbStatusCountController struct {
	sqlDB *sql.DB
}

func NewDbStatusCountController(sqlDB *sql.DB) *DbStatusCountController {
	return &DbStatusCountController{sqlDB: sqlDB}
}

type dbStatusCountAPIResponse struct {
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    *dbStatusCountData  `json:"data,omitempty"`
}

type dbStatusCountData struct {
	Table  string `json:"table"`
	Status int    `json:"status"`
	Count  int64  `json:"count"`
}

type dbStatusCountRequest struct {
	Table string `json:"table"`
	Tabel string `json:"tabel"` // 兼容拼写
	Status *int  `json:"status"`
}

func writeDbStatusCountJSON(w http.ResponseWriter, code int, message string, data *dbStatusCountData) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dbStatusCountAPIResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// StatusCount GET/POST
// GET  /api/db/status-count?table=bilibili_video&status=2
// POST /api/db/status-count  body: {"table":"bilibili_video","status":2}
func (c *DbStatusCountController) StatusCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if c.sqlDB == nil {
		writeDbStatusCountJSON(w, 500, "database not available", nil)
		return
	}

	table, statusVal, ok := c.parseStatusCountInput(w, r)
	if !ok {
		return
	}

	realTable, allowed := allowedStatusCountTables[strings.ToLower(table)]
	if !allowed {
		writeDbStatusCountJSON(w, 400, "unsupported table: "+table, nil)
		return
	}

	q := "SELECT COUNT(*) FROM `" + realTable + "` WHERE status = ?"
	var count int64
	if err := c.sqlDB.QueryRow(q, statusVal).Scan(&count); err != nil {
		writeDbStatusCountJSON(w, 500, "query failed: "+err.Error(), nil)
		return
	}

	data := &dbStatusCountData{
		Table:  realTable,
		Status: statusVal,
		Count:  count,
	}
	msg := fmt.Sprintf("table=%s status=%d 满足条件的记录数 count=%d", realTable, statusVal, count)
	writeDbStatusCountJSON(w, 200, msg, data)
}

func (c *DbStatusCountController) parseStatusCountInput(w http.ResponseWriter, r *http.Request) (table string, status int, ok bool) {
	if r.Method == http.MethodGet {
		table = strings.TrimSpace(r.URL.Query().Get("table"))
		if table == "" {
			table = strings.TrimSpace(r.URL.Query().Get("tabel"))
		}
		if table == "" {
			writeDbStatusCountJSON(w, 400, "table is required", nil)
			return "", 0, false
		}
		s := strings.TrimSpace(r.URL.Query().Get("status"))
		if s == "" {
			writeDbStatusCountJSON(w, 400, "status is required", nil)
			return "", 0, false
		}
		v, err := strconv.Atoi(s)
		if err != nil {
			writeDbStatusCountJSON(w, 400, "invalid status", nil)
			return "", 0, false
		}
		return table, v, true
	}

	var req dbStatusCountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDbStatusCountJSON(w, 400, "invalid json body", nil)
		return "", 0, false
	}
	table = strings.TrimSpace(req.Table)
	if table == "" {
		table = strings.TrimSpace(req.Tabel)
	}
	if table == "" {
		writeDbStatusCountJSON(w, 400, "table is required", nil)
		return "", 0, false
	}
	if req.Status == nil {
		writeDbStatusCountJSON(w, 400, "status is required", nil)
		return "", 0, false
	}
	return table, *req.Status, true
}
