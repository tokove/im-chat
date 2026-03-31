# 面试考察点与参考答案

> 基于 im-chat 项目的技术知识点整理，涵盖 Go 语言、实时通信、数据库、身份认证等方向。

---

## 一、Go 语言基础

**Q1：Go 的 Goroutine 和线程有什么区别？**

> **参考答案**：
> - Goroutine 由 Go 运行时调度，是用户态的轻量级并发单元，初始栈仅 2KB，可按需增长；OS 线程由操作系统调度，栈固定（通常 1~8MB）。
> - Go 使用 M:N 调度模型（多个 Goroutine 对应多个 OS 线程），调度开销远低于线程上下文切换。
> - 在本项目中，每个 WebSocket 连接启动 3 个 goroutine（readPump / writePump / heartbeat），即使有数千个并发连接，内存占用也远小于线程模型。

---

**Q2：如何理解 Channel？有缓冲和无缓冲 Channel 的区别是什么？**

> **参考答案**：
> - Channel 是 goroutine 间通信的管道，遵循"共享内存通过通信"的原则。
> - **无缓冲**：发送方会阻塞，直到接收方接收，实现同步通信。
> - **有缓冲**：发送方在缓冲区未满时不阻塞，实现异步通信。
> - 本项目中每个客户端的 `send` channel 缓冲为 512，防止慢消费者（如网络抖动的客户端）阻塞其他用户的消息投递。

---

**Q3：sync.RWMutex 和 sync.Mutex 的区别？项目中为什么用 RWMutex？**

> **参考答案**：
> - `Mutex` 互斥锁：读写都会阻塞其他所有操作。
> - `RWMutex` 读写锁：允许多个 goroutine 同时持有读锁，写锁会阻塞所有读/写操作。
> - Hub 中存储在线用户 Map，**查找连接（读操作）**远比**注册/注销（写操作）**频繁，使用 RWMutex 可以提高并发吞吐量。

---

**Q4：如何实现 Goroutine 的优雅退出？**

> **参考答案**：
> - 使用 `context.WithCancel` 或 `context.WithTimeout`，关闭时调用 `cancel()` 通知 goroutine 退出。
> - 使用关闭的 channel 作为退出信号（`close(quit)`）。
> - 本项目中通过监听 `SIGINT/SIGTERM` 信号触发服务器关闭，并通过 Hub 向所有 WebSocket 客户端广播 `EventShutdown` 事件，让客户端主动断开连接。

---

**Q5：什么是 panic 和 recover？项目中如何使用？**

> **参考答案**：
> - `panic` 触发运行时错误并开始展开调用栈；`recover` 在 `defer` 中捕获 panic，阻止程序崩溃。
> - 在 WebSocket 的 readPump 中使用 `defer recover()`，防止单个连接的异常导致整个服务器崩溃。

---

## 二、WebSocket 与实时通信

**Q6：HTTP 和 WebSocket 的区别是什么？**

> **参考答案**：
> - HTTP 是请求-响应模型，每次通信需要建立连接，通信完成后关闭（或 Keep-Alive 复用）。
> - WebSocket 在 HTTP 握手后升级协议，建立持久全双工连接，双方可随时主动推送数据，延迟更低，适合实时场景。
> - WebSocket 升级请求：请求头包含 `Upgrade: websocket` 和 `Connection: Upgrade`，服务端响应 `101 Switching Protocols`。

---

**Q7：描述一下本项目的 WebSocket 消息投递流程。**

> **参考答案**：
> 1. 发送方 WebSocket 连接发送 `message` 事件（JSON）
> 2. readPump 读取事件，解析目标用户 ID
> 3. 消息存入 SQLite（状态 `created`）
> 4. Hub 查找目标用户的 `*Client`
>    - **在线**：通过 `client.send` channel 推送消息；接收方 writePump 发送给客户端，同时更新消息状态为 `delivered`，反向推送 `EventDelivered` 给发送方
>    - **离线**：消息保留 `created` 状态，等待下次上线时 Hub 批量重推

---

**Q8：如何保证 WebSocket 连接的健康状态？**

> **参考答案**：
> - 服务端每 30 秒发送 Ping 帧，客户端需在 5 秒内回复 Pong 帧。
> - 若超时未收到 Pong，认为连接断开，关闭 WebSocket 连接并从 Hub 注销该客户端。
> - 这是 WebSocket 的标准心跳保活机制，防止连接因网络中断而变成"僵尸连接"。

---

**Q9：如果系统需要水平扩展到多台服务器，当前的 Hub 模式会有什么问题？如何解决？**

> **参考答案**：
> - 问题：Hub 是进程内 in-memory 存储，不同服务器实例之间无法共享连接状态，A 服务器上的用户无法收到 B 服务器上的用户发来的消息。
> - 解决方案：
>   - 引入 **Redis Pub/Sub** 或 **消息队列（Kafka/RabbitMQ）** 作为跨节点消息总线
>   - 将用户连接映射存储到 **Redis**（`userID → serverID`），路由消息到正确节点
>   - 或使用 **一致性哈希**，将同一用户的连接路由到固定节点

---

## 三、数据库

**Q10：什么是 SQLite 的 WAL 模式？为什么要使用它？**

> **参考答案**：
> - WAL（Write-Ahead Logging）是一种日志模式，写操作先写入 WAL 文件，读操作可直接读取主数据库文件，读写不互斥。
> - 相比默认的 DELETE 模式，WAL 模式允许多个读操作与一个写操作并发执行，显著提升并发读写性能，适合 WebSocket 消息频繁写入的场景。

---

**Q11：如何设计防止私聊会话重复创建？**

> **参考答案**：
> - 在数据库层对 `(user1_id, user2_id)` 设置 `UNIQUE` 约束。
> - 在代码层强制 `user1_id < user2_id`（字符串比较），确保两个用户之间的会话无论谁发起都映射到同一行记录，避免 `(A,B)` 和 `(B,A)` 两条记录并存。

---

**Q12：分页查询如何实现？有什么优化方案？**

> **参考答案**：
> - 当前实现：`LIMIT ? OFFSET ?`，按 `created_at DESC` 排序。
> - 问题：深分页时 OFFSET 越大性能越差（需扫描跳过的行）。
> - 优化方案：**游标分页（Cursor-based Pagination）**，用最后一条记录的 `id` 或 `created_at` 作为游标，下次查询 `WHERE created_at < :cursor LIMIT ?`，性能稳定 O(log n)。

---

## 四、身份认证

**Q13：JWT 的工作原理是什么？有什么优缺点？**

> **参考答案**：
> - JWT = Header（算法）+ Payload（Claims）+ Signature，三部分 Base64 编码后以 `.` 拼接。
> - 服务端用私钥（HMAC-SHA256）对 Header+Payload 签名，验证时重新计算签名比对即可，**无需查询数据库**，天然支持水平扩展。
> - 优点：无状态、跨服务、减少 DB 查询。
> - 缺点：Token 无法主动失效（需配合黑名单或短有效期），Payload 可被解码（不应存储敏感信息）。

---

**Q14：Access Token + Refresh Token 双 Token 方案的意义是什么？**

> **参考答案**：
> - Access Token 有效期短（本项目 2 小时），即使泄露，危害窗口小。
> - Refresh Token 有效期长，存储在数据库中，**可以主动撤销**（登出时清空），弥补 JWT 无法失效的缺陷。
> - 本项目按平台分别存储 Refresh Token，单平台登出只影响该平台，不影响其他平台。

---

**Q15：bcrypt 哈希和 MD5/SHA 哈希有什么区别？为什么密码要用 bcrypt？**

> **参考答案**：
> - MD5/SHA 是快速哈希，专为数据完整性校验设计，暴力破解很快（GPU 每秒数十亿次）。
> - bcrypt 是自适应慢哈希，内置加盐（salt）防彩虹表，`cost` 参数控制计算开销，可随硬件提升调高。
> - 密码存储必须使用慢哈希（bcrypt/argon2/scrypt），不能使用 MD5/SHA。

---

## 五、架构 & 系统设计

**Q16：如果消息量增大，SQLite 撑不住了，怎么迁移到 MySQL/PostgreSQL？**

> **参考答案**：
> - 项目使用标准 `database/sql` 接口 + 预编译语句，切换数据库驱动（`go-sql-driver/mysql` 或 `lib/pq`）并调整少量 SQL 语法（如自增主键、时间函数）即可迁移。
> - 数据库操作集中在 `internal/model/` 层，改动范围可控，体现了良好的分层设计。

---

**Q17：项目中文件上传有哪些安全考虑？**

> **参考答案**：
> 1. **MIME 类型白名单**：只允许特定类型文件，拒绝可执行文件上传（防 WebShell）。
> 2. **文件大小限制**：50MB 上限，防止 DoS 攻击。
> 3. **UUID 文件名**：不使用原始文件名，防止路径遍历攻击（`../../etc/passwd`）。
> 4. **鉴权保护**：上传接口需要 JWT 认证，且只能上传到自己参与的会话目录。

---

**Q18：项目的优雅关闭是如何实现的？**

> **参考答案**：
> 1. 使用 `signal.NotifyContext` 监听 `SIGINT` / `SIGTERM` 信号。
> 2. 收到信号后，Hub 向所有在线 WebSocket 客户端推送 `EventShutdown` 事件，通知客户端断开并重连。
> 3. 调用 `http.Server.Shutdown(ctx)` 关闭 HTTP 服务器，设置最长 10 秒等待时间，让正在处理的请求自然完成后退出。

---

## 六、开放性问题

**Q19：如果让你对这个项目做性能优化，你会从哪些方面入手？**

> **参考答案（思路）**：
> - **消息队列**：引入 Redis / Kafka 解耦消息写入和推送，提高吞吐量。
> - **连接池**：SQLite 换成 MySQL/PgSQL 并配置连接池参数（`SetMaxOpenConns` / `SetMaxIdleConns`）。
> - **缓存**：用 Redis 缓存用户信息、会话信息，减少 DB 查询。
> - **压缩**：WebSocket 开启 permessage-deflate 压缩，减少带宽占用。
> - **分库分表**：消息表按时间或用户 ID 水平分片，应对海量消息。

---

**Q20：如果让你给这个项目加单元测试，你会怎么做？**

> **参考答案**：
> - 对 `pkg/utils/` 中的 JWT 生成/验证、bcrypt 加密函数编写纯函数单元测试。
> - 对 `internal/model/` 中的数据库操作，使用 `testing` 包 + 内存 SQLite（`":memory:"`）做集成测试。
> - 对 HTTP 接口，使用 `net/http/httptest` 包模拟请求，结合 Gin 的测试模式。
> - 对 WebSocket Hub 逻辑，Mock `*Client` 接口，测试消息路由和状态流转。
