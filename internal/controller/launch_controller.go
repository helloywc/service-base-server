package controller

import (
	"encoding/json"
	"net/http"
	"regexp"

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

// Bootstrap 执行 launchctl bootstrap
// POST /api/bootstrap/{name}  例如 name=mysql-dev
func (c *LaunchController) Bootstrap(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
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
		Message: "bootstrap ok",
		Stdout:  res.Stdout, Stderr: res.Stderr,
	})
}

// Bootout 执行 launchctl bootout
// POST /api/bootout/{name}  例如 name=mysql-dev
func (c *LaunchController) Bootout(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
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
		Message: "bootout ok",
		Stdout:  res.Stdout, Stderr: res.Stderr,
	})
}
