package view

// LaunchResponse 统一响应（View 层），含命令终端输出；成功时 code=200
type LaunchResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Stdout  string `json:"stdout,omitempty"` // 命令标准输出
	Stderr  string `json:"stderr,omitempty"` // 命令标准错误
}

// ErrorResponse 错误响应，含命令终端输出；以 HTTP 状态码判断（如 400/500）
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Stdout  string `json:"stdout,omitempty"`
	Stderr  string `json:"stderr,omitempty"`
}
