package view

// LaunchResponse 统一响应（View 层），含命令终端输出；成功时 code=200；始终带 stdout、stderr
type LaunchResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Stdout  string   `json:"stdout"` // 命令标准输出
	Stderr  string   `json:"stderr"` // 命令标准错误
	Files   []string `json:"files,omitempty"` // 仅 archive 列表接口使用；其他接口不返回此字段
}

// ErrorResponse 错误响应，含命令终端输出；以 HTTP 状态码判断（如 400/500）；始终带 stdout、stderr
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Stdout  string `json:"stdout"`
	Stderr  string `json:"stderr"`
}

// DeleteArchivesResponse 批量删除 zip 的响应
type DeleteArchivesResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Deleted []string `json:"deleted"` // 已删除的文件名（无 .zip）
	Failed  []string `json:"failed"`  // 删除失败的文件名及原因，如 "file: not found"
}
