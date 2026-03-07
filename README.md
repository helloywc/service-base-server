# code-server

基于 Go 的 HTTP 后端服务。

## 要求

- Go 1.21+

## 项目结构（MVC）

```
.
├── cmd/server/              # 程序入口
├── internal/
│   ├── model/               # Model：领域（如 LaunchService）
│   │   └── launch.go
│   ├── view/                # View：请求/响应 DTO
│   │   └── launch.go
│   ├── controller/          # Controller：bootstrap / bootout 接口
│   │   └── launch_controller.go
│   ├── service/             # 业务：执行 launchctl 命令
│   │   └── launchctl.go
│   ├── handler/             # 通用 handler（health 等）
│   └── server/              # 路由与服务器配置
├── go.mod
└── README.md
```

## 运行

```bash
# 默认端口 8080
go run ./cmd/server

# 指定端口
PORT=3000 go run ./cmd/server
```

## 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/` | 欢迎信息 |
| GET | `/health` | 健康检查（探活） |
| POST | `/api/bootstrap/{name}` | 执行 `launchctl bootstrap gui/$(id -u) .../name.plist`，如 `mysql-dev` |
| POST | `/api/bootout/{name}` | 执行 `launchctl bootout gui/$(id -u) .../name.plist`，如 `mysql-dev` |

`name` 仅允许字母、数字、`_`、`.`、`-`。plist 路径：`/Users/wilson1/Library/LaunchAgents/{name}.plist`。

## 构建

```bash
go build -o bin/server ./cmd/server
./bin/server
```

## 优雅关闭

服务监听 `SIGINT` / `SIGTERM`，收到信号后会等待最多 10 秒完成当前请求再退出。
