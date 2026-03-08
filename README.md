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
# 开发环境（默认）：端口 8080
go run ./cmd/server
# 或显式
APP_ENV=dev go run ./cmd/server

# 生产环境：端口 8090
APP_ENV=prod go run ./cmd/server

# 自定义端口（覆盖 APP_ENV 的默认端口）
PORT=3000 go run ./cmd/server
```

**环境变量**

| 变量 | 说明 | 默认 |
|------|------|------|
| `APP_ENV` | `dev` → 8080，`prod` → 8090 | `dev`（8080） |
| `PORT` | 监听端口，设置时优先于 `APP_ENV` | 见上 |

## 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/` | 欢迎信息 |
| GET | `/health` | 健康检查（探活） |
| POST | `/api/bootstrap/{name}` | 执行命令 `launchctl bootstrap ...`，如 `mysql-dev`；成功返回 HTTP 200 且 body 含 `code: 200` |
| POST | `/api/bootout/{name}` | 执行命令 `launchctl bootout ...`，如 `mysql-dev`；成功返回 HTTP 200 且 body 含 `code: 200` |
| GET | `/api/list/{name}` | 查询状态（`launchctl list | grep name`），如 `mysql-dev`；成功返回 HTTP 200 且 body 含 `code: 200` |
| POST | `/api/archive/{name}` | 将 name 对应目录打 zip，zip 放在同级，命名为 `name_YYYY-MM-DD_HH-mm-ss.zip`；成功返回 HTTP 200，`stdout` 为 zip 绝对路径 |
| GET | `/api/archive/{name}` | 列出以 `name_` 或 `name-` 开头的文件，返回 `files` 数组（仅文件名、无 .zip），按日期倒序 |
| POST | `/api/extract/{name}/{timestamp}` | 解压对应 zip 到 zip 所在目录；timestamp 格式 `YYYY-MM-DD_HH-mm-ss`，如 `mysql-dev` + `2026-03-08_07-16-48` |
| POST | `/api/archives/delete` | Body 为 JSON 数组，如 `["mysql-dev_2026-03-08_17-44-45", ...]`，按文件名删除对应 zip |

`name` 仅允许字母、数字、`_`、`.`、`-`。plist 路径：`/Users/wilson1/Library/LaunchAgents/{name}.plist`。

---

## 接口调用归纳

以下假设服务地址为 `http://localhost:8080`（dev）。prod 为 8090，或使用 `http://<ip>:<port>`。

### 通用

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 欢迎 | GET | `/` | 返回服务说明 |
| 健康检查 | GET | `/health` | 探活，返回 `{"status":"ok"}` |

```bash
curl http://localhost:8080/
curl http://localhost:8080/health
```

---

### LaunchAgent（launchctl）

| 接口 | 方法 | 路径 | 传参 | 说明 |
|------|------|------|------|------|
| 启动服务 | POST | `/api/bootstrap/{name}` | 路径中 `name`，如 `mysql-dev` | 执行 `launchctl bootstrap`，加载对应 plist |
| 停止服务 | POST | `/api/bootout/{name}` | 同上 | 执行 `launchctl bootout` |
| 查询状态 | GET | `/api/list/{name}` | 同上 | 执行 `launchctl list \| grep name`，返回匹配行 |

```bash
# 启动 mysql-dev 对应的 LaunchAgent
curl -X POST http://localhost:8080/api/bootstrap/mysql-dev

# 停止
curl -X POST http://localhost:8080/api/bootout/mysql-dev

# 查询是否在运行
curl http://localhost:8080/api/list/mysql-dev
```

成功响应均为 HTTP 200，body 含 `code`、`message`、`stdout`、`stderr`（命令输出在 `stdout`/`stderr`）。

---

### 压缩与解压（目录：`/Users/yang/Operator/Databases/`，name 中 `-` 对应路径 `/`）

| 接口 | 方法 | 路径 | 传参 | 说明 |
|------|------|------|------|------|
| 打 zip 包 | POST | `/api/archive/{name}` | 路径中 `name`，如 `mysql-dev` | 将 `.../mysql/dev` 打 zip 到同级目录，文件名 `name_YYYY-MM-DD_HH-mm-ss.zip` |
| 列 zip 列表 | GET | `/api/archive/{name}` | 同上 | 返回以 `name_` 或 `name-` 开头的文件名列表（无路径、无 .zip），按时间倒序，在 `files` 数组 |
| 解压 | POST | `/api/extract/{name}/{timestamp}` | 路径中 `name` + `timestamp` | 解压对应 zip 到 zip 所在目录；timestamp 格式 `YYYY-MM-DD_HH-mm-ss` |
| 批量删除 zip | POST | `/api/archives/delete` | Body：JSON 数组，元素为文件名（无 .zip） | 删除对应 zip；返回 `deleted`、`failed` 列表 |

```bash
# 打包 mysql-dev 对应目录
curl -X POST http://localhost:8080/api/archive/mysql-dev

# 列出该 name 下的 zip（仅文件名，无 .zip 后缀）
curl http://localhost:8080/api/archive/mysql-dev

# 解压指定时间戳的 zip 到当前目录
curl -X POST http://localhost:8080/api/extract/mysql-dev/2026-03-08_07-16-48

# 批量删除 zip（body 为文件名数组，无 .zip）
curl -X POST http://localhost:8080/api/archives/delete \
  -H "Content-Type: application/json" \
  -d '["mysql-dev_2026-03-08_17-44-45","mysql-dev_2026-03-08_15-00-07"]'
```

- 打包成功：HTTP 200，`stdout` 为 zip 的绝对路径。
- 列表成功：HTTP 200，`files` 为字符串数组，如 `["mysql-dev-2026-03-08_06-58-52", ...]`。
- 解压成功：HTTP 200，`message": "extract ok"`。
- 批量删除成功：HTTP 200，body 含 `deleted`（已删除文件名数组）、`failed`（失败项及原因数组）；部分失败时 `message` 为 `"partial"`。

---

### 错误与约定

- 参数错误：HTTP 400，body 含 `code`、`message`。
- 方法不允许：HTTP 405（如对 bootstrap 发 GET）。
- 服务端错误：HTTP 500，`message` 为错误信息。
- 成功一律 HTTP 200，且 body 中 `code: 200`（列表接口还有 `files`）。

## 构建

```bash
go build -o bin/server ./cmd/server
./bin/server
```

## 优雅关闭

服务监听 `SIGINT` / `SIGTERM`，收到信号后会等待最多 10 秒完成当前请求再退出。
