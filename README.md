# code-server

基于 Go 的 HTTP 后端服务。

## 要求

- Go 1.21+

## 项目结构

```
.
├── cmd/
│   └── server/          # 程序入口
│       └── main.go
├── internal/
│   ├── handler/         # HTTP 处理函数
│   │   └── handler.go
│   └── server/          # 服务器配置与路由
│       └── server.go
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

## 构建

```bash
go build -o bin/server ./cmd/server
./bin/server
```

## 优雅关闭

服务监听 `SIGINT` / `SIGTERM`，收到信号后会等待最多 10 秒完成当前请求再退出。
