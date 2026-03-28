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

支持用 **`.env`** 配置（类似 Node.js 的 dotenv），启动时自动加载项目根目录下的：

1. **`.env`** — 全环境通用基线（进程里已存在的环境变量不会被文件覆盖）
2. **`.env.development` / `.env.test` / `.env.production`** — 由 `APP_ENV` 决定加载哪一个（未设置 `APP_ENV` 时默认按 `development` 加载 `.env.development`）
3. **`.env.local`** — 本机最后覆盖（可选，勿提交）

```bash
make env-init          # 从 .env*.example 生成缺失文件
cp .env.example .env   # 若尚未创建基线

# 开发：显式设置 APP_ENV 后启动（端口默认 8080）
make dev
# 或
APP_ENV=development go run ./cmd/server

# 测试 / 生产
make test-env
make prod
# 或 APP_ENV=test / APP_ENV=production go run ./cmd/server
```

`.env` 中可写 `APP_ENV=development`（或 `test` / `production`），以便只执行 `go run ./cmd/server` 时也能加载对应环境文件。

也可直接使用环境变量，优先级高于 `.env`：

```bash
PORT=3000 go run ./cmd/server
```

**环境变量**

| 变量 | 说明 | 默认 |
|------|------|------|
| `APP_ENV` | `development`/`test` → 8080，`production` → 8090 | `development`（8080） |
| `PORT` | 监听端口，设置时优先于 `APP_ENV` | 见上 |
| `MEILISEARCH_HOST` | Meilisearch 地址 | `http://localhost:7700` |
| `MEILISEARCH_API_KEY` | Meilisearch API Key | `123456` |

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

### Meilisearch（索引与文档，供前端调用）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/meili/indexes` | 索引列表 |
| POST | `/api/meili/indexes` | 创建索引，body `{"uid":"movies","primaryKey":"id"}` |
| GET | `/api/meili/indexes/:uid` | 获取单个索引 |
| PUT | `/api/meili/indexes/:uid` | 更新索引，body `{"primaryKey":"id"}` |
| DELETE | `/api/meili/indexes/:uid` | 删除索引 |
| PUT | `/api/meili/indexes/:uid/documents` | 添加/替换文档，body 为对象数组 `[{...}]` |
| GET | `/api/meili/indexes/:uid/documents` | 文档列表，query `limit`、`offset` |
| GET | `/api/meili/indexes/:uid/documents/:documentId` | 获取单条文档 |
| DELETE | `/api/meili/indexes/:uid/documents/:documentId` | 删除单条文档 |
| POST | `/api/meili/indexes/:uid/documents/delete-batch` | 批量删除，body `{"ids":["1","2"]}` |
| DELETE | `/api/meili/indexes/:uid/documents` | 删除该索引下全部文档 |

`name` 仅允许字母、数字、`_`、`.`、`-`。plist 路径：`/Users/wilson1/Library/LaunchAgents/{name}.plist`。

---

## 接口调用归纳

以下假设服务地址为 `http://localhost:8080`（development）。prod 为 8090，或使用 `http://<ip>:<port>`。

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

### Meilisearch（索引与文档）

需本地运行 Meilisearch（如 `http://localhost:7700`），并设置 `MEILISEARCH_HOST`、`MEILISEARCH_API_KEY`（默认 key 为 `123456`）。响应格式与 Meilisearch 官方 API 一致。

```bash
# 索引：列表、创建、获取、更新、删除
curl http://localhost:8080/api/meili/indexes
curl -X POST http://localhost:8080/api/meili/indexes -H "Content-Type: application/json" -d '{"uid":"movies","primaryKey":"id"}'
curl http://localhost:8080/api/meili/indexes/movies
curl -X PUT http://localhost:8080/api/meili/indexes/movies -H "Content-Type: application/json" -d '{"primaryKey":"id"}'
curl -X DELETE http://localhost:8080/api/meili/indexes/movies

# 文档：添加、列表、获取一条、删除一条、批量删除、清空
curl -X PUT http://localhost:8080/api/meili/indexes/movies/documents -H "Content-Type: application/json" -d '[{"id":1,"title":"Hello"}]'
curl "http://localhost:8080/api/meili/indexes/movies/documents?limit=10&offset=0"
curl http://localhost:8080/api/meili/indexes/movies/documents/1
curl -X DELETE http://localhost:8080/api/meili/indexes/movies/documents/1
curl -X POST http://localhost:8080/api/meili/indexes/movies/documents/delete-batch -H "Content-Type: application/json" -d '{"ids":["1","2"]}'
curl -X DELETE http://localhost:8080/api/meili/indexes/movies/documents
```

---

### 错误与约定

- 参数错误：HTTP 400，body 含 `code`、`message`。
- 方法不允许：HTTP 405（如对 bootstrap 发 GET）。
- 服务端错误：HTTP 500，`message` 为错误信息。
- 成功一律 HTTP 200，且 body 中 `code: 200`（列表接口还有 `files`）。

---

## curl 示例汇总

以下将 `BASE` 设为 `http://localhost:8080`，远程可改为 `http://192.168.2.101:8080` 等。按需复制执行。

```bash
BASE=http://localhost:8080
```

### 通用

```bash
# 欢迎页
curl $BASE/

# 健康检查
curl $BASE/health
```

### LaunchAgent（launchctl）

```bash
# 启动服务（如 mysql-dev）
curl -X POST $BASE/api/bootstrap/mysql-dev

# 停止服务
curl -X POST $BASE/api/bootout/mysql-dev

# 查询状态（是否在运行）
curl $BASE/api/list/mysql-dev
```

### 压缩与解压

```bash
# 打包：将 mysql-dev 对应目录打 zip
curl -X POST $BASE/api/archive/mysql-dev

# 列表：列出该 name 下的 zip 文件名（无 .zip）
curl $BASE/api/archive/mysql-dev

# 解压：按时间戳解压到当前目录
curl -X POST $BASE/api/extract/mysql-dev/2026-03-08_07-16-48

# 批量删除 zip（body 为文件名数组）
curl -X POST $BASE/api/archives/delete \
  -H "Content-Type: application/json" \
  -d '["mysql-dev_2026-03-08_17-44-45","mysql-dev_2026-03-08_15-00-07"]'
```

### Meilisearch 索引

```bash
# 索引列表
curl $BASE/api/meili/indexes

# 创建索引（uid + 可选 primaryKey）
curl -X POST $BASE/api/meili/indexes \
  -H "Content-Type: application/json" \
  -d '{"uid":"movies","primaryKey":"id"}'

# 获取单个索引
curl $BASE/api/meili/indexes/movies

# 更新索引（如改 primaryKey）
curl -X PUT $BASE/api/meili/indexes/movies \
  -H "Content-Type: application/json" \
  -d '{"primaryKey":"id"}'

# 删除索引
curl -X DELETE $BASE/api/meili/indexes/movies
```

### Meilisearch 文档

```bash
# 添加/替换文档（body 为对象数组）
curl -X PUT $BASE/api/meili/indexes/movies/documents \
  -H "Content-Type: application/json" \
  -d '[{"id":1,"title":"Hello World"},{"id":2,"title":"Foo"}]'

# 文档列表（分页）
curl "$BASE/api/meili/indexes/movies/documents?limit=10&offset=0"

# 获取单条文档
curl $BASE/api/meili/indexes/movies/documents/1

# 删除单条文档
curl -X DELETE $BASE/api/meili/indexes/movies/documents/1

# 批量删除文档
curl -X POST $BASE/api/meili/indexes/movies/documents/delete-batch \
  -H "Content-Type: application/json" \
  -d '{"ids":["1","2","3"]}'

# 删除该索引下全部文档
curl -X DELETE $BASE/api/meili/indexes/movies/documents
```

## 构建

```bash
go build -o bin/server ./cmd/server
./bin/server
```

## 优雅关闭

服务监听 `SIGINT` / `SIGTERM`，收到信号后会等待最多 10 秒完成当前请求再退出。
