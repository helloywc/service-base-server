package view

// IndexCreateRequest 创建索引请求
type IndexCreateRequest struct {
	UID        string `json:"uid"`
	PrimaryKey string `json:"primaryKey,omitempty"`
}

// MeiliErrorResponse Meilisearch 相关错误响应
type MeiliErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Body    string `json:"body,omitempty"` // 原始 Meilisearch 错误 body
}
