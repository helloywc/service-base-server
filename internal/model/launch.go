package model

// LaunchService 表示一个 LaunchAgent 服务（用于 bootstrap/bootout）
type LaunchService struct {
	Name string // 服务名，对应 plist 文件名（不含 .plist），如 mysql-dev
}
