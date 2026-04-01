# im-chat · 即时通讯聊天系统

基于 **Go + WebSocket** 构建的轻量级即时通讯后端服务，支持私聊、实时消息收发、文件分享、用户在线状态感知及消息送达/已读回执。

---

## 目录

- [功能特性](#功能特性)
- [技术栈](#技术栈)
- [架构概览](#架构概览)
- [快速开始](#快速开始)
- [Flutter 前端](#flutter-前端)
- [环境变量](#环境变量)
- [项目结构](#项目结构)
- [API 文档](#api-文档)
- [WebSocket 事件](#websocket-事件)
- [数据库结构](#数据库结构)

---

## 功能特性

- **用户系统**：邮箱注册与登录，bcrypt 密码哈希，JWT 双 Token 鉴权
- **私聊会话**：创建并管理一对一私聊会话
- **实时消息**：基于 WebSocket 的双向实时消息收发
- **消息回执**：送达回执（delivered）与已读回执（read）双向通知
- **离线消息**：用户重新上线后自动补发所有未送达消息
- **在线状态**：用户上线/下线事件实时广播给所有连接的客户端
- **正在输入**：typing 指示器事件支持
- **文件上传**：支持图片、文档、媒体文件上传（最大 50 MB）
- **分页消息历史**：分页查询历史消息，支持前端无限滚动
- **多平台 Token**：Web 与 Mobile 平台各自独立的 Refresh Token
- **优雅关闭**：捕获系统信号后，先通知所有客户端再安全关闭服务

---

## 技术栈

| 分类 | 技术 |
|------|------|
| 语言 | Go 1.25 |
| HTTP 框架 | [Gin](https://github.com/gin-gonic/gin) v1.12 |
| WebSocket | [coder/websocket](https://github.com/coder/websocket) v1.8 |
| 数据库 | SQLite（[modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite)）|
| 认证 | JWT ([golang-jwt/jwt](https://github.com/golang-jwt/jwt)) + bcrypt |
| 唯一 ID | [google/uuid](https://github.com/google/uuid) |
| 配置 | [cleanenv](https://github.com/ilyakaznacheev/cleanenv) |

---

## 架构概览

```
┌─────────────────────────────────────────────────────────┐
│                   Gin HTTP 服务器 :8080                  │
├──────────────────────────┬──────────────────────────────┤
│      REST API 路由        │       WebSocket 路由          │
│  (公开 + JWT 保护路由)     │    /api/ws  (JWT 认证)        │
└──────────────────────────┴──────────┬──────────────────-┘
                                      │
                           ┌──────────▼──────────┐
                           │   Hub（连接管理器）   │
                           │  sync.RWMutex 保护   │
                           │  userID → *Client    │
                           └──────────┬──────────┘
                                      │ channel 事件传递
                           ┌──────────▼──────────┐
                           │     SQLite 数据库     │
                           │  WAL 模式 · 外键约束  │
                           └─────────────────────┘
```

**连接模型**：每个 WebSocket 连接独立运行 3 个 goroutine（readPump / writePump / heartbeat），通过缓冲大小为 512 的 channel 异步投递消息，Hub 使用 `sync.RWMutex` 保证并发安全。

---

## 快速开始

### 前置条件

- Go 1.25 或更高版本
- Git

### 克隆并运行

```bash
# 克隆仓库
git clone https://github.com/tokove/im-chat.git
cd im-chat/backend

# 下载依赖
go mod download

# 创建配置文件目录及示例配置（可选）
mkdir -p config
cat > config/dev.env <<EOF
ENV=dev
DB_PATH=sqlite/dev
DB_NAME=api.db
HTTP_ADDRESS=localhost:8080
JWT_KEY=your-secret-jwt-key
EOF

# 启动服务
go run ./cmd/api/ -config ./config/dev.env
```

服务启动后，终端会输出所有可用接口地址：

```
Server is running: http://localhost:8080
Health Check HTTP, GET: http://localhost:8080/api/health-check-http
Health Check Websocket, GET: ws://localhost:8080/api/health-check-ws
...
```

### 使用 rest.http 测试

项目根目录提供了 [`rest.http`](./backend/rest.http) 文件，可在支持 HTTP Client 的编辑器（如 VS Code REST Client 插件）中直接执行接口测试。

---

## Flutter 前端

Flutter 前端位于 `mobileapp/` 目录，支持 Android、iOS、macOS、Linux、Windows 及 **Web** 平台。

### 前置条件

- Flutter SDK（包含 Dart）
- 后端服务已启动（参见[快速开始](#快速开始)）

### 安装依赖

```bash
cd im-chat/mobileapp
flutter pub get
```

### 在浏览器中运行（推荐用于多用户测试）

```bash
flutter run -d web-server --web-hostname 0.0.0.0 --web-port 3000
```

服务启动后，在浏览器中打开 `http://localhost:3000`。打开多个浏览器窗口，分别注册或登录不同账号，即可实时体验双向聊天。

### 在移动设备或桌面上运行

```bash
# 列出可用设备
flutter devices

# 在指定设备上运行
flutter run -d <device-id>
```

---

## 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `ENV` | `dev` | 运行环境（dev / prod） |
| `DB_PATH` | `sqlite/dev` | SQLite 数据库目录 |
| `DB_NAME` | `api.db` | SQLite 数据库文件名 |
| `HTTP_ADDRESS` | `localhost:8080` | HTTP 服务监听地址 |
| `JWT_KEY` | `supersecretjwtkey` | JWT 签名密钥（**生产环境请务必修改**） |
| `CONFIG_PATH` | — | .env 配置文件路径（也可通过 `-config` 参数传入） |

---

## 项目结构

```
backend/
├── cmd/
│   └── api/
│       └── main.go              # 应用入口：依赖注入、优雅关闭
├── config/
│   └── dev.env                  # 本地开发配置文件（不提交到版本控制）
├── internal/
│   ├── config/
│   │   └── config.go            # 配置加载（cleanenv）
│   ├── db/
│   │   └── db.go                # 数据库初始化、建表、索引
│   ├── middleware/
│   │   ├── authenticate.go      # JWT 认证中间件
│   │   └── cors.go              # CORS 中间件
│   ├── model/
│   │   ├── user.go              # 用户数据模型 & 数据库操作
│   │   ├── message.go           # 消息模型 & CRUD
│   │   └── private.go           # 私聊会话模型 & 查询
│   ├── realtime/
│   │   ├── hub.go               # WebSocket 连接管理器（Hub）
│   │   ├── client.go            # 单个 WebSocket 连接（Client）
│   │   └── event.go             # 事件类型定义
│   └── router/
│       ├── router.go            # 路由注册
│       ├── auth.go              # 认证相关接口
│       ├── conversation.go      # 会话相关接口
│       ├── websocket.go         # WebSocket 升级处理
│       ├── file.go              # 文件上传/下载
│       ├── user.go              # 用户相关接口
│       └── health.go            # 健康检查接口
├── pkg/
│   ├── response/
│   │   └── response.go          # 统一 JSON 响应格式
│   └── utils/
│       ├── jwt.go               # JWT 生成与解析
│       └── crypto.go            # 密码哈希工具
├── go.mod
├── go.sum
└── rest.http                    # HTTP 接口测试文件
```

---

## API 文档

### 统一响应格式

```json
{
  "code": 200,
  "success": true,
  "message": "success",
  "data": {}
}
```

### 认证说明

受保护接口需在请求头中携带：

```
Authorization: Bearer <access_token>
X-Platform: web | mobile
```

---

### 健康检查

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/health-check-http` | HTTP 健康检查 |
| `GET` | `/api/health-check-ws` | WebSocket 健康检查（echo 回显） |

---

### 认证接口

#### 注册

```
POST /api/auth/register-email
```

请求体：

```json
{
  "name": "张三",
  "email": "zhangsan@example.com",
  "password": "12345678"
}
```

#### 登录

```
POST /api/auth/login-email
X-Platform: web | mobile
```

请求体：

```json
{
  "email": "zhangsan@example.com",
  "password": "12345678"
}
```

响应 `data`：

```json
{
  "access_token": "<jwt>",
  "refresh_token": "<token>"
}
```

#### 刷新 Token

```
POST /api/auth/refresh-session
X-Platform: web | mobile
```

请求体：

```json
{ "refresh_token": "<refresh_token>" }
```

#### 登出

```
POST /api/auth/logout  （需认证）
```

#### 获取当前用户

```
POST /api/auth/current-user  （需认证）
```

---

### 用户接口

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/users/:id` | 根据 ID 获取用户信息（需认证） |

---

### 会话接口

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/conversations` | 获取当前用户的所有私聊会话（需认证） |
| `GET` | `/api/conversations/privates/:private_id` | 获取指定私聊会话详情（需认证） |
| `POST` | `/api/conversations/privates/create` | 创建新的私聊会话（需认证） |
| `GET` | `/api/conversations/privates/:private_id/messages` | 分页获取会话消息（需认证） |

消息分页参数：`?page=1&limit=20`（limit 最大 100）

---

### 文件接口

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/files/:private_id` | 上传文件（form-data，字段名 `file`，需认证） |
| `GET` | `/api/files/*filepath` | 下载/访问文件（需认证） |

**上传限制**：
- 最大文件大小：50 MB
- 允许类型：`.jpg` `.jpeg` `.png` `.gif` `.pdf` `.doc` `.docx` `.txt` `.mp4` `.mov` `.mp3`

---

### WebSocket

```
GET /api/ws
Authorization: Bearer <access_token>
```

连接成功后，服务端会立即推送 `current_users` 事件，告知当前在线用户列表。

---

## WebSocket 事件

所有事件格式：

```json
{
  "event_type": "<事件类型>",
  "payload": {}
}
```

### 客户端 → 服务端

| 事件类型 | 说明 | payload 示例 |
|----------|------|--------------|
| `message` | 发送消息 | `{ "to_id": 2, "private_id": 1, "message_type": "text", "content": "你好" }` |
| `read` | 标记消息已读 | `{ "message_id": 42, "from_id": 2 }` |
| `typing` | 正在输入指示 | `{ "to_id": 2, "private_id": 1 }` |

### 服务端 → 客户端

| 事件类型 | 触发时机 | payload 说明 |
|----------|----------|--------------|
| `current_users` | 连接建立时 | 当前在线用户 ID 列表 |
| `online` | 有用户上线 | 上线用户的信息 |
| `offline` | 有用户下线 | 下线用户的 ID |
| `message` | 收到新消息 | 完整消息对象 |
| `delivered` | 消息送达接收方 | `{ "message_id": 42 }` |
| `read` | 消息被接收方已读 | `{ "message_id": 42 }` |
| `typing` | 对方正在输入 | `{ "from_id": 2, "private_id": 1 }` |
| `new_private` | 新私聊会话被创建 | 会话对象 |
| `heartbeat` | 每 30 秒一次 | — |
| `error` | 发生错误 | `{ "message": "错误描述" }` |
| `shutdown` | 服务器即将关闭 | — |

---

## 数据库结构

```sql
-- 用户表
users (id, name, email, password,
       refresh_token_web, refresh_token_web_at,
       refresh_token_mobile, refresh_token_mobile_at,
       created_at)

-- 私聊会话表（user1_id < user2_id 保证唯一性）
privates (id, user1_id, user2_id, created_at)
  UNIQUE(user1_id, user2_id)
  CHECK(user1_id < user2_id)

-- 消息表
messages (id, from_id, private_id, message_type,
          content, delivered, read, created_at)
```

SQLite PRAGMA 配置：`WAL` 模式、外键约束、5s 忙等待超时、`NORMAL` 同步模式。

---

## License

[MIT](LICENSE)
