# im-chat 项目知识点总结

## 一、技术栈概览

| 分类 | 技术 | 版本 |
|------|------|------|
| 编程语言 | Go | 1.25.0 |
| HTTP 框架 | Gin | v1.12.0 |
| WebSocket | coder/websocket | v1.8.14 |
| 数据库 | SQLite (modernc) | v1.46.1 |
| 身份认证 | JWT (golang-jwt/jwt) | v5.3.1 |
| 密码哈希 | bcrypt (golang.org/x/crypto) | v0.48.0 |
| 唯一标识 | UUID (google/uuid) | v1.6.0 |
| 配置管理 | cleanenv | v1.5.0 |

---

## 二、架构设计知识点

### 2.1 分层架构

```
cmd/api/main.go            ← 应用入口，依赖注入，优雅关闭
internal/router/           ← HTTP 路由层（Controller 层）
internal/model/            ← 数据模型 + 数据库查询（Model 层）
internal/realtime/         ← WebSocket 实时通信层
internal/middleware/        ← 中间件（认证、CORS）
internal/db/               ← 数据库初始化、建表
internal/config/           ← 配置加载
pkg/utils/                 ← 公共工具（JWT、加密）
pkg/response/              ← 统一响应格式
```

### 2.2 实时通信架构（Hub 模式）

```
WebSocket 客户端 A ──┐
WebSocket 客户端 B ──┤──→ Hub（中央连接管理器）──→ SQLite
WebSocket 客户端 N ──┘       ↕ channel 传递事件
```

- **Hub**：维护在线用户 Map（`userID → *Client`），使用 `sync.RWMutex` 保证并发安全
- **Client**：每个 WebSocket 连接独立的读/写/心跳 goroutine
- **Channel**：缓冲大小 512 的事件队列，实现异步非阻塞消息投递

---

## 三、Go 语言核心知识点

### 3.1 并发编程

- **Goroutine**：每个 WebSocket 连接启动 3 个 goroutine（readPump、writePump、heartbeat）
- **Channel**：有缓冲 channel 用于事件队列（`make(chan Event, 512)`），防止慢消费者阻塞
- **sync.RWMutex**：读多写少场景（Hub 查找连接用读锁，注册/注销用写锁）
- **context.Context**：用于 goroutine 的优雅取消和超时控制
- **panic/recover**：在 readPump 中捕获异常，保证连接安全关闭

### 3.2 标准库使用

- `database/sql`：通过接口操作 SQLite，预编译语句（Prepared Statement）防 SQL 注入
- `encoding/json`：WebSocket 事件的序列化与反序列化
- `net/http`：HTTP 服务器，配合 Gin 框架
- `os`：文件目录操作（文件上传存储路径）
- `time`：心跳超时（30s Ping，5s Pong 超时）

### 3.3 错误处理

- 函数返回 `(value, error)` 二元组
- HTTP 层统一返回结构化 JSON 错误
- WebSocket 层使用 `EventError` 事件推送错误给客户端
- 不使用 `panic` 作为正常流程控制

---

## 四、WebSocket 实时通信知识点

### 4.1 连接生命周期

```
HTTP 升级请求 (Upgrade: websocket)
    ↓
JWT 认证通过
    ↓
注册到 Hub（广播上线事件）
    ↓
启动 readPump + writePump + heartbeat goroutine
    ↓
消息收发循环
    ↓
连接断开 → 从 Hub 注销（广播下线事件）
```

### 4.2 事件类型系统

| 事件名 | 方向 | 说明 |
|--------|------|------|
| `message` | 客户端→服务端 | 发送消息 |
| `delivered` | 服务端→客户端 | 消息已送达回执 |
| `read` | 客户端→服务端 | 标记消息已读 |
| `typing` | 客户端→服务端 | 正在输入指示 |
| `online` | 服务端→客户端 | 用户上线通知 |
| `offline` | 服务端→客户端 | 用户下线通知 |
| `current_users` | 服务端→客户端 | 当前在线用户列表 |
| `heartbeat` | 服务端→客户端 | 心跳包（保活） |
| `error` | 服务端→客户端 | 错误通知 |
| `shutdown` | 服务端→客户端 | 服务关闭通知 |

### 4.3 消息投递保证

- **三态模型**：`created` → `delivered` → `read`
- **断线重连恢复**：用户重新上线时，Hub 查询所有未送达消息并重新推送
- **发送方回执**：`EventDelivered`、`EventRead` 反向推送给消息发送方

---

## 五、数据库设计知识点

### 5.1 表结构

```sql
-- 用户表
users (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT NOT NULL UNIQUE,
  password TEXT NOT NULL,
  refresh_token_web TEXT,
  refresh_token_mobile TEXT,
  created_at DATETIME,
  updated_at DATETIME
)

-- 私聊会话表
privates (
  id TEXT PRIMARY KEY,
  user1_id TEXT REFERENCES users(id),
  user2_id TEXT REFERENCES users(id),
  created_at DATETIME,
  UNIQUE(user1_id, user2_id)  -- 唯一约束，user1_id < user2_id
)

-- 消息表
messages (
  id TEXT PRIMARY KEY,
  from_id TEXT REFERENCES users(id),
  private_id TEXT REFERENCES privates(id),
  message_type TEXT,  -- text / file
  content TEXT,
  delivered INTEGER DEFAULT 0,
  read INTEGER DEFAULT 0,
  created_at DATETIME
)
```

### 5.2 SQLite 性能优化

```sql
PRAGMA journal_mode = WAL;         -- 写前日志，支持并发读写
PRAGMA foreign_keys = ON;          -- 启用外键约束
PRAGMA busy_timeout = 5000;        -- 锁等待超时 5 秒
PRAGMA synchronous = NORMAL;       -- 平衡性能与安全
```

- 对 `messages.private_id`、`messages.from_id`、`messages.created_at` 建立索引

### 5.3 分页查询模式

```sql
SELECT * FROM messages
WHERE private_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?
```

返回 `has_next_page` 字段用于前端分页控制。

---

## 六、身份认证知识点

### 6.1 JWT 认证流程

```
注册 → bcrypt 哈希密码存储
登录 → 验证密码 → 生成 Access Token（2小时）+ Refresh Token（存数据库）
请求 → Bearer Token 认证 → JWT 校验（签名+过期时间）
刷新 → Refresh Token 换新 Access Token + 新 Refresh Token（轮换）
登出 → 清空数据库中的 Refresh Token
```

### 6.2 多平台 Token 管理

- `X-Platform` 请求头区分 web / mobile
- 每个平台维护独立的 `refresh_token_web` / `refresh_token_mobile`
- 防止跨平台 Token 复用

### 6.3 JWT Claims 结构

```go
type Claims struct {
    UserID   string
    Name     string
    Platform string  // "web" | "mobile"
    jwt.RegisteredClaims  // 包含 ExpiresAt 等标准字段
}
```

---

## 七、文件上传知识点

### 7.1 文件处理逻辑

- **大小限制**：50MB
- **类型白名单**：图片（jpg/png/gif）、文档（pdf/doc/docx/txt）、媒体（mp4/mov/mp3）
- **存储路径**：`files/chats/{private_id}/{sender_id}/{uuid}.ext`
- **静态服务**：使用 Gin 的 `Static` 中间件提供文件下载

### 7.2 安全考虑

- 通过 MIME 类型白名单过滤恶意文件
- UUID 命名文件，防止路径遍历攻击
- 只有会话参与者才能上传到对应的 private_id 目录

---

## 八、API 设计知识点

### 8.1 RESTful 设计规范

- 资源命名使用复数名词（`/conversations`、`/messages`）
- HTTP 动词语义化（GET 查询、POST 创建）
- 分层路径（`/conversations/privates/:id/messages`）

### 8.2 统一响应格式

```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

### 8.3 中间件链

```
CORS → JWT 认证 → 平台校验 → 路由处理器
```

---

## 九、工程化知识点

### 9.1 优雅关闭

```go
// 监听 SIGINT / SIGTERM 信号
// 关闭前广播 shutdown 事件给所有 WebSocket 客户端
// 等待最长 10 秒让连接自然关闭
```

### 9.2 配置管理

- 环境变量 + `.env` 文件双模式
- 使用 `cleanenv` 进行结构体标签绑定
- 默认值注入（`env-default:"dev"`）

### 9.3 项目结构最佳实践

- `internal/` 包：限制外部引用，保护内部实现
- `pkg/` 包：可被外部复用的公共工具
- `cmd/` 包：应用入口，只做组装，不含业务逻辑

---

## 十、知识点思维导图

```
im-chat
├── 实时通信
│   ├── WebSocket 协议升级
│   ├── Hub 模式（连接管理）
│   ├── 事件驱动架构
│   └── 消息三态（created/delivered/read）
├── Go 并发
│   ├── Goroutine 生命周期管理
│   ├── Channel 异步通信
│   ├── sync.RWMutex 读写锁
│   └── Context 超时取消
├── 身份认证
│   ├── JWT（无状态认证）
│   ├── bcrypt 密码哈希
│   ├── Refresh Token 轮换
│   └── 多平台 Token 隔离
├── 数据库
│   ├── SQLite WAL 模式
│   ├── 外键约束
│   ├── 索引优化
│   └── 分页查询
├── HTTP 框架
│   ├── Gin 路由与中间件
│   ├── RESTful API 设计
│   ├── CORS 跨域处理
│   └── 文件上传下载
└── 工程实践
    ├── 分层架构
    ├── 优雅关闭
    ├── 配置管理
    └── 错误处理
```
