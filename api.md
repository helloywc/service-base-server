# code-server API 文档

本文档描述 code-server 提供的 HTTP 接口。服务默认地址：**开发** `http://localhost:8080`，**生产** `http://localhost:8090`。

---

## 约定

- **Content-Type**：请求体为 JSON 时使用 `Content-Type: application/json`。
- **成功**：HTTP 状态码 200，业务接口 body 中通常含 `code: 200`。
- **错误**：4xx/5xx，body 含 `code`、`message`，部分含 `stdout`/`stderr` 或 `body`（原始错误）。
- **CORS**：已开放，支持跨域调用。
- **路径参数**：`{name}` 仅允许字母、数字、`_`、`.`、`-`。

---

## 一、通用

### 欢迎页

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/` | 返回服务说明与文档链接 |

**curl**

```bash
curl http://localhost:8080/
```

---

### 健康检查

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 探活，返回 `{"status":"ok"}` |

**curl**

```bash
curl http://localhost:8080/health
```

---

## 二、LaunchAgent（launchctl）

路径中的 `{name}` 对应 plist 名（如 `mysql-dev` → `/Users/wilson1/Library/LaunchAgents/mysql-dev.plist`）。

### 启动服务

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/bootstrap/{name}` | 执行 `launchctl bootstrap gui/$(id -u) .../name.plist` |

**响应**：`code`、`message`、`stdout`、`stderr`（命令输出）。

**curl**

```bash
curl -X POST http://localhost:8080/api/bootstrap/mysql-dev
```

---

### 停止服务

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/bootout/{name}` | 执行 `launchctl bootout gui/$(id -u) .../name.plist` |

**curl**

```bash
curl -X POST http://localhost:8080/api/bootout/mysql-dev
```

---

### 查询状态

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/list/{name}` | 执行 `launchctl list | grep name`，返回匹配行（在 `stdout`） |

**curl**

```bash
curl http://localhost:8080/api/list/mysql-dev
```

---

## 三、压缩与解压

目录规则：`name` 中 `-` 映射为路径 `/`，基础目录 `/Users/yang/Operator/Databases/`。  
例如 `mysql-dev` → 目录 `.../mysql/dev`，zip 放在 `.../mysql/`。

### 打 zip 包

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/archive/{name}` | 将对应目录打 zip，命名为 `name_YYYY-MM-DD_HH-mm-ss.zip`，放在同级目录 |

**响应**：`code`、`message`、`stdout`（zip 绝对路径）、`stderr`。

**curl**

```bash
curl -X POST http://localhost:8080/api/archive/mysql-dev
```

---

### 列 zip 列表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/archive/{name}` | 列出以 `name_` 或 `name-` 开头的文件名（无路径、无 .zip），按时间倒序 |

**响应**：`code`、`message`、`files`（字符串数组）。

**curl**

```bash
curl http://localhost:8080/api/archive/mysql-dev
```

---

### 解压

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/extract/{name}/{timestamp}` | 解压对应 zip 到 zip 所在目录。timestamp 格式 `YYYY-MM-DD_HH-mm-ss` |

**curl**

```bash
curl -X POST http://localhost:8080/api/extract/mysql-dev/2026-03-08_07-16-48
```

---

### 批量删除 zip

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/archives/delete` | Body 为文件名数组（无 .zip），删除对应 zip |

**请求体示例**

```json
["mysql-dev_2026-03-08_17-44-45", "mysql-dev_2026-03-08_15-00-07"]
```

**响应**：`code`、`message`、`deleted`（已删除）、`failed`（失败及原因）。部分失败时 `message` 为 `"partial"`。

**curl**

```bash
curl -X POST http://localhost:8080/api/archives/delete \
  -H "Content-Type: application/json" \
  -d '["mysql-dev_2026-03-08_17-44-45","mysql-dev_2026-03-08_15-00-07"]'
```

---

## 四、Meilisearch 索引

需配置 `MEILISEARCH_HOST`（默认 `http://localhost:7700`）、`MEILISEARCH_API_KEY`（默认 `123456`）。响应格式与 [Meilisearch 官方 API](https://docs.meilisearch.com/reference/api/overview) 一致。

### 索引列表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/meili/indexes` | 获取所有索引 |

**curl**

```bash
curl http://localhost:8080/api/meili/indexes
```

---

### 创建索引

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/meili/indexes` | 创建索引 |

**请求体**

```json
{
  "uid": "movies",
  "primaryKey": "id"
}
```

**curl**

```bash
curl -X POST http://localhost:8080/api/meili/indexes \
  -H "Content-Type: application/json" \
  -d '{"uid":"movies","primaryKey":"id"}'
```

---

### 获取单个索引

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/meili/indexes/:uid` | 获取指定索引信息 |

**curl**

```bash
curl http://localhost:8080/api/meili/indexes/movies
```

---

### 更新索引

| 方法 | 路径 | 说明 |
|------|------|------|
| PUT | `/api/meili/indexes/:uid` | 更新索引（如 primaryKey） |

**请求体示例**

```json
{ "primaryKey": "id" }
```

**curl**

```bash
curl -X PUT http://localhost:8080/api/meili/indexes/movies \
  -H "Content-Type: application/json" \
  -d '{"primaryKey":"id"}'
```

---

### 删除索引

| 方法 | 路径 | 说明 |
|------|------|------|
| DELETE | `/api/meili/indexes/:uid` | 删除索引 |

**curl**

```bash
curl -X DELETE http://localhost:8080/api/meili/indexes/movies
```

---

## 五、Meilisearch 文档

### 添加/替换文档

| 方法 | 路径 | 说明 |
|------|------|------|
| PUT | `/api/meili/indexes/:uid/documents` | 添加或替换文档，body 为对象数组 |

**请求体示例**

```json
[
  { "id": 1, "title": "Hello World" },
  { "id": 2, "title": "Foo" }
]
```

**curl**

```bash
curl -X PUT http://localhost:8080/api/meili/indexes/movies/documents \
  -H "Content-Type: application/json" \
  -d '[{"id":1,"title":"Hello World"},{"id":2,"title":"Foo"}]'
```

---

### 文档列表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/meili/indexes/:uid/documents` | 文档列表，支持 query `limit`、`offset` |

**curl**

```bash
curl "http://localhost:8080/api/meili/indexes/movies/documents?limit=10&offset=0"
```

---

### 获取单条文档

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/meili/indexes/:uid/documents/:documentId` | 获取指定文档 |

**curl**

```bash
curl http://localhost:8080/api/meili/indexes/movies/documents/1
```

---

### 删除单条文档

| 方法 | 路径 | 说明 |
|------|------|------|
| DELETE | `/api/meili/indexes/:uid/documents/:documentId` | 删除指定文档 |

**curl**

```bash
curl -X DELETE http://localhost:8080/api/meili/indexes/movies/documents/1
```

---

### 批量删除文档

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/meili/indexes/:uid/documents/delete-batch` | 按 id 批量删除 |

**请求体**

```json
{ "ids": ["1", "2", "3"] }
```

**curl**

```bash
curl -X POST http://localhost:8080/api/meili/indexes/movies/documents/delete-batch \
  -H "Content-Type: application/json" \
  -d '{"ids":["1","2","3"]}'
```

---

### 删除全部文档

| 方法 | 路径 | 说明 |
|------|------|------|
| DELETE | `/api/meili/indexes/:uid/documents` | 删除该索引下全部文档 |

**curl**

```bash
curl -X DELETE http://localhost:8080/api/meili/indexes/movies/documents
```

---

## 六、错误码

| HTTP 状态 | 说明 |
|-----------|------|
| 200 | 成功 |
| 400 | 参数错误（如缺少 uid、非法 name） |
| 405 | 方法不允许（如对 bootstrap 使用 GET） |
| 404 | 资源不存在（如 Meilisearch 路径错误） |
| 500 | 服务端错误（如命令执行失败、Meilisearch 返回错误） |

错误响应 body 示例：

```json
{
  "code": 400,
  "message": "invalid or missing name (use only letters, numbers, _, ., -)",
  "stdout": "",
  "stderr": ""
}
```

Meilisearch 相关错误可能包含 `body` 字段（Meilisearch 原始响应）。
