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

## 六、Deepseek AI 接口

### 内容验证 / 通用大模型调用

后端封装了对 Deepseek `/chat/completions` 的代理接口，前端只需调用本服务即可，无需直接暴露 API Key。

需在环境变量中配置（由后端读取并拼接到请求头中）：

- `DEEPSEEK_ADDRESS`（例如 `https://api.deepseek.com`）
- `DEEPSEEK_API_KEY`（用于生成 `Authorization: Bearer <DEEPSEEK_API_KEY>` 请求头）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/deepseek` | 调用 Deepseek Chat，以 JSON 结果形式返回 |

**请求体**

```json
{
  "prompt": "system prompt，用于约束模型角色与输出格式",
  "text": "用户输入内容"
}
```

- `prompt`：会作为 `role: "system"` 的消息传递给 Deepseek。
- `text`：会作为 `role: "user"` 的消息传递给 Deepseek（也可传 `content`，与 `text` 二选一，优先 `text`）。

内部实际请求 Deepseek 时等价于：

**Headers**

```http
Authorization: Bearer <DEEPSEEK_API_KEY>
Content-Type: application/json
```

**Body**

```json
{
  "model": "deepseek-chat",
  "messages": [
    { "role": "system", "content": "<prompt>" },
    { "role": "user", "content": "<text 或 content>" }
  ],
  "stream": false
}
```

**响应**

- 成功：状态码 200，Body 为 Deepseek 原始 JSON，结构示例：`id`、`object`、`model`、`choices`（其中 `choices[0].message.content` 为助手回复正文）、`usage` 等。
- 失败：Deepseek 非 2xx 时返回 `502` 并透传错误内容；配置缺失时返回 `500`。

**curl 示例**

```bash
curl -X POST http://localhost:8080/api/deepseek \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "请为我校对这段语句的错别字",
    "text": "鸡旦灌饼是什么黑暗料理？老板怕不是把鸡蛋写成母鸡下蛋的现场了叭"
  }'
```

---

### Deepseek 校验任务（DB 驱动）/deepseek-verify

该组接口会：

- **从表 `ai_agent_text`** 读取 `key` 对应的 `content` 作为 `messages[0].content`（system）
- **从表 `bilibili_video`** 读取 `status=1` 且 `priority` 最大的一条的 `context` 作为 `messages[1].content`（user）
- 调用 Deepseek `https://api.deepseek.com/chat/completions`，并解析 `choices[0].message.content`
- **成功**：将该条 `bilibili_video.context` 更新为模型返回 content，并将 `status=2`
- **失败**：将该条 `bilibili_video.status=-2`

#### 1) 启动持续校验任务 `/api/deepseek-verify/start`

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/deepseek-verify/start` | 启动后台循环：无论成功/失败都会继续处理下一条，直到没有 `status=1` 的记录自动停止 |

**curl**

```bash
curl -X POST http://localhost:8080/api/deepseek-verify/start \
  -H "Content-Type: application/json" \
  -d '{
    "key": "text-verify",
    "table": "bilibili_video"
  }'
```

**响应示例**

```json
{ "running": true, "key": "text-verify" }
```

#### 2) 停止持续校验任务 `/api/deepseek-verify/stop`

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/deepseek-verify/stop` | 终止后台持续校验任务 |

**curl**

```bash
curl -X POST http://localhost:8080/api/deepseek-verify/stop
```

**响应示例**

```json
{ "running": false }
```

#### 3) 查询任务状态 `/api/deepseek-verify/status`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/deepseek-verify/status` | 查询后台持续校验任务是否运行 |

**curl**

```bash
curl http://localhost:8080/api/deepseek-verify/status
```

**响应示例**

```json
{ "running": true, "key": "text-verify" }
```

（若未运行则 `running=false`，`key` 可能为空。）

## 七、错误码

| HTTP 状态 | 说明 |
|-----------|------|
| 200 | 成功 |
| 400 | 参数错误（如缺少 uid、非法 name） |
| 405 | 方法不允许（如对 bootstrap 使用 GET） |
| 404 | 资源不存在（如 Meilisearch 路径错误） |
| 500 | 服务端错误（如命令执行失败、Meilisearch / Deepseek 调用失败等） |

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
