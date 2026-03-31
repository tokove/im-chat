# 个人简历

## 基本信息

| | |
|---|---|
| **姓名** | [你的姓名] |
| **求职意向** | Go 后端开发工程师 |
| **工作年限** | [X] 年 |
| **联系方式** | [邮箱] / [手机] |
| **GitHub** | [https://github.com/tokove/im-chat](https://github.com/tokove/im-chat) |

---

## 技能清单

**后端开发**
- 熟练掌握 **Go 语言**，熟悉并发编程（Goroutine、Channel、sync 包）
- 熟练使用 **Gin** 框架进行 RESTful API 开发
- 熟悉 **WebSocket** 协议，有实时通信系统开发经验
- 熟练掌握 **JWT** 鉴权机制，了解 Token 刷新与多平台隔离方案
- 熟悉 **SQLite / MySQL**，了解索引优化、WAL 模式、分页查询等数据库实践

**架构 & 设计**
- 了解事件驱动架构、Hub 连接管理模式
- 具备分层架构（MVC/Clean Architecture）设计能力
- 熟悉 RESTful API 设计规范与统一响应格式封装

**工程化**
- 熟悉 Go Modules 依赖管理
- 了解优雅启停（Graceful Shutdown）、配置中心（环境变量 + .env 文件）
- 具备基础的 Linux 运维能力，了解容器化部署

---

## 项目经历

### im-chat — 即时通讯聊天系统

**项目描述**

基于 Go 语言开发的即时通讯后端服务，支持用户注册登录、私聊会话管理、实时消息收发、文件上传分享等核心功能。

**技术栈**

`Go 1.25` · `Gin` · `WebSocket (coder/websocket)` · `SQLite` · `JWT` · `bcrypt`

**核心职责与成果**

- **实时消息系统**：设计并实现基于 **Hub 模式**的 WebSocket 连接管理器，使用 `sync.RWMutex` 保证多 goroutine 并发安全，每个连接独立维护读/写/心跳三个 goroutine，支持消息**送达回执**和**已读回执**双向通知。

- **消息可靠投递**：实现消息三态（`created` → `delivered` → `read`）状态机，用户断线重连后自动推送所有未送达消息，保证消息不丢失。

- **鉴权系统**：实现基于 JWT 的无状态认证方案，Access Token（2小时有效期）+ Refresh Token（数据库持久化）双 Token 架构，支持 Web / Mobile 双平台 Token 独立管理，防止跨平台 Token 复用。

- **数据库设计**：使用 SQLite WAL 模式支持高并发读写，对消息表的 `private_id`、`from_id`、`created_at` 字段建立索引，实现分页查询，返回 `has_next_page` 字段辅助前端分页加载。

- **文件服务**：实现文件上传接口，支持 50MB 限制、MIME 类型白名单校验（图片/文档/媒体），使用 UUID 命名防路径遍历，按 `会话ID/发送者ID/` 目录结构存储。

- **工程实践**：实现优雅关闭（捕获 SIGINT/SIGTERM 信号，广播 shutdown 事件，等待连接自然关闭），使用 `cleanenv` 管理多环境配置，规范 `internal/`、`pkg/`、`cmd/` 目录结构。

**项目地址**：[https://github.com/tokove/im-chat](https://github.com/tokove/im-chat)

---

## 教育背景

| 学校 | 专业 | 学历 | 时间 |
|------|------|------|------|
| [学校名称] | [计算机科学与技术 / 软件工程] | 本科 / 硕士 | [入学年份] — [毕业年份] |

---

## 自我评价

具备扎实的 Go 语言基础，对并发编程、实时通信系统有实际项目经验。注重代码质量与工程规范，善于分层架构设计。学习能力强，能快速上手新技术栈，有良好的问题分析与解决能力。
