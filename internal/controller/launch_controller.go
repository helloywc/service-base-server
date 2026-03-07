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
