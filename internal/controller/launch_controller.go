package controller

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"code-server/internal/service"
	"code-server/internal/view"
)

// 服务名只允许字母、数字、下划线、连字符，防止路径注入
var safeName = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

// LaunchController bootstrap/bootout 接口（Controller 层）
type LaunchController struct {
	launch *service.LaunchCtl
}

// NewLaunchController 创建控制器
func NewLaunchController(launch *service.LaunchCtl) *LaunchController {
	return &LaunchController{launch: launch}
}

func (c *LaunchController) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (c *LaunchController) validateName(name string) bool {
	return name != "" && safeName.MatchString(name)
}

// nameFromPath 从路径中解析服务名，兼容 Go 1.21 无 PathValue 时使用
func nameFromPath(path, prefix string) string {
	s := strings.TrimPrefix(path, prefix)
	s = strings.Trim(s, "/")
	if i := strings.Index(s, "/"); i >= 0 {
		s = s[:i]
	}
	return s
}

// nameAndTimestampFromPath 从路径解析 name 与 timestamp，如 /api/extract/mysql-dev/2026-03-08_07-16-48
func nameAndTimestampFromPath(path, prefix string) (name, timestamp string, ok bool) {
	s := strings.TrimPrefix(path, prefix)
	s = strings.Trim(s, "/")
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// Bootstrap 执行 launchctl bootstrap（仅 POST）
// POST /api/bootstrap/{name}  例如 /api/bootstrap/mysql-dev
func (c *LaunchController) Bootstrap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.writeJSON(w, http.StatusMethodNotAllowed, view.ErrorResponse{Code: 405, Message: "method not allowed, use POST"})
		return
	}
	name := nameFromPath(r.URL.Path, "/api/bootstrap/")
	if !c.validateName(name) {
		c.writeJSON(w, http.StatusBadRequest, view.ErrorResponse{Code: 400, Message: "invalid or missing name (use only letters, numbers, _, ., -)"})
		return
	}
	res, err := c.launch.Bootstrap(name)
	if err != nil {
		c.writeJSON(w, http.StatusInternalServerError, view.ErrorResponse{
			Code: 500, Message: err.Error(),
			Stdout: res.Stdout, Stderr: res.Stderr,
		})
		return
	}
	c.writeJSON(w, http.StatusOK, view.LaunchResponse{
		Code: 200, Message: "bootstrap ok",
		Stdout: res.Stdout, Stderr: res.Stderr,
	})
}

// Bootout 执行 launchctl bootout（仅 POST）
// POST /api/bootout/{name}  例如 /api/bootout/mysql-dev
func (c *LaunchController) Bootout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.writeJSON(w, http.StatusMethodNotAllowed, view.ErrorResponse{Code: 405, Message: "method not allowed, use POST"})
		return
	}
	name := nameFromPath(r.URL.Path, "/api/bootout/")
	if !c.validateName(name) {
		c.writeJSON(w, http.StatusBadRequest, view.ErrorResponse{Code: 400, Message: "invalid or missing name (use only letters, numbers, _, ., -)"})
		return
	}
	res, err := c.launch.Bootout(name)
	if err != nil {
		c.writeJSON(w, http.StatusInternalServerError, view.ErrorResponse{
			Code: 500, Message: err.Error(),
			Stdout: res.Stdout, Stderr: res.Stderr,
		})
		return
	}
	c.writeJSON(w, http.StatusOK, view.LaunchResponse{
		Code: 200, Message: "bootout ok",
		Stdout: res.Stdout, Stderr: res.Stderr,
	})
}

// List 查询 launchctl list | grep name（仅 GET）
// GET /api/list/{name}  例如 /api/list/mysql-dev
func (c *LaunchController) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.writeJSON(w, http.StatusMethodNotAllowed, view.ErrorResponse{Code: 405, Message: "method not allowed, use GET"})
		return
	}
	name := nameFromPath(r.URL.Path, "/api/list/")
	if !c.validateName(name) {
		c.writeJSON(w, http.StatusBadRequest, view.ErrorResponse{Code: 400, Message: "invalid or missing name (use only letters, numbers, _, ., -)"})
		return
	}
	res, err := c.launch.List(name)
	if err != nil {
		c.writeJSON(w, http.StatusInternalServerError, view.ErrorResponse{
			Code: 500, Message: err.Error(),
			Stdout: res.Stdout, Stderr: res.Stderr,
		})
		return
	}
	c.writeJSON(w, http.StatusOK, view.LaunchResponse{
		Code: 200, Message: "list ok",
		Stdout: res.Stdout, Stderr: res.Stderr,
	})
}

// ArchiveList 列出以 name_ 开头的文件（仅 GET）
// GET /api/archive/{name}  例如 /api/archive/mysql-dev -> 返回 mysql-dev_*.zip 等列表（stdout 每行一个完整路径）
func (c *LaunchController) ArchiveList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.writeJSON(w, http.StatusMethodNotAllowed, view.ErrorResponse{Code: 405, Message: "method not allowed, use GET", Stdout: "", Stderr: ""})
		return
	}
	name := nameFromPath(r.URL.Path, "/api/archive/")
	if !c.validateName(name) {
		c.writeJSON(w, http.StatusBadRequest, view.ErrorResponse{Code: 400, Message: "invalid or missing name (use only letters, numbers, _, ., -)", Stdout: "", Stderr: ""})
		return
	}
	files, err := service.ListArchiveFiles(name)
	if err != nil {
		c.writeJSON(w, http.StatusInternalServerError, view.ErrorResponse{Code: 500, Message: err.Error(), Stdout: "", Stderr: ""})
		return
	}
	if files == nil {
		files = []string{}
	}
	c.writeJSON(w, http.StatusOK, view.LaunchResponse{
		Code: 200, Message: "ok",
		Stdout: "", Stderr: "",
		Files:  files,
	})
}

// Archive 将 name 对应目录打 zip 包（仅 POST）；GET 时由 ArchiveList 处理
// POST /api/archive/{name}
func (c *LaunchController) Archive(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		c.ArchiveList(w, r)
		return
	}
	if r.Method != http.MethodPost {
		c.writeJSON(w, http.StatusMethodNotAllowed, view.ErrorResponse{Code: 405, Message: "method not allowed, use GET or POST", Stdout: "", Stderr: ""})
		return
	}
	name := nameFromPath(r.URL.Path, "/api/archive/")
	if !c.validateName(name) {
		c.writeJSON(w, http.StatusBadRequest, view.ErrorResponse{Code: 400, Message: "invalid or missing name (use only letters, numbers, _, ., -)", Stdout: "", Stderr: ""})
		return
	}
	zipPath, err := service.Archive(name)
	if err != nil {
		c.writeJSON(w, http.StatusInternalServerError, view.ErrorResponse{Code: 500, Message: err.Error(), Stdout: "", Stderr: ""})
		return
	}
	c.writeJSON(w, http.StatusOK, view.LaunchResponse{
		Code: 200, Message: "archive ok",
		Stdout: zipPath, Stderr: "",
	})
}

// 时间戳格式：YYYY-MM-DD_HH-mm-ss
var timestampRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}$`)

// Extract 解压 name 对应的 zip（timestamp 如 2026-03-08_07-16-48）到 zip 所在目录（仅 POST）
// POST /api/extract/{name}/{timestamp}  例如 /api/extract/mysql-dev/2026-03-08_07-16-48
func (c *LaunchController) Extract(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.writeJSON(w, http.StatusMethodNotAllowed, view.ErrorResponse{Code: 405, Message: "method not allowed, use POST", Stdout: "", Stderr: ""})
		return
	}
	name, timestamp, ok := nameAndTimestampFromPath(r.URL.Path, "/api/extract/")
	if !ok {
		c.writeJSON(w, http.StatusBadRequest, view.ErrorResponse{Code: 400, Message: "invalid path, use /api/extract/{name}/{timestamp}", Stdout: "", Stderr: ""})
		return
	}
	if !c.validateName(name) {
		c.writeJSON(w, http.StatusBadRequest, view.ErrorResponse{Code: 400, Message: "invalid or missing name (use only letters, numbers, _, ., -)", Stdout: "", Stderr: ""})
		return
	}
	if !timestampRegex.MatchString(timestamp) {
		c.writeJSON(w, http.StatusBadRequest, view.ErrorResponse{Code: 400, Message: "invalid timestamp, use YYYY-MM-DD_HH-mm-ss", Stdout: "", Stderr: ""})
		return
	}
	if err := service.Extract(name, timestamp); err != nil {
		c.writeJSON(w, http.StatusInternalServerError, view.ErrorResponse{Code: 500, Message: err.Error(), Stdout: "", Stderr: ""})
		return
	}
	c.writeJSON(w, http.StatusOK, view.LaunchResponse{
		Code: 200, Message: "extract ok",
		Stdout: "", Stderr: "",
	})
}

// 文件名（无 .zip）允许字母、数字、下划线、连字符
var safeArchiveFilename = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

// DeleteArchives 批量删除 zip（仅 POST）；body 为 JSON 数组，如 ["mysql-dev_2026-03-08_17-44-45", ...]
// POST /api/archives/delete
func (c *LaunchController) DeleteArchives(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.writeJSON(w, http.StatusMethodNotAllowed, view.ErrorResponse{Code: 405, Message: "method not allowed, use POST", Stdout: "", Stderr: ""})
		return
	}
	var filenames []string
	if err := json.NewDecoder(r.Body).Decode(&filenames); err != nil {
		c.writeJSON(w, http.StatusBadRequest, view.ErrorResponse{Code: 400, Message: "invalid body, expect JSON array of strings", Stdout: "", Stderr: ""})
		return
	}
	var valid []string
	for _, f := range filenames {
		f = strings.TrimSpace(f)
		if f != "" && safeArchiveFilename.MatchString(f) {
			valid = append(valid, f)
		}
	}
	deleted, failed := service.DeleteArchiveFiles(valid)
	msg := "delete ok"
	if len(failed) > 0 {
		msg = "partial"
	}
	c.writeJSON(w, http.StatusOK, view.DeleteArchivesResponse{
		Code: 200, Message: msg,
		Deleted: deleted, Failed: failed,
	})
}
