# AuroraIM — 即时通讯系统

> 基于 Go + Vue 3 构建的全栈即时通讯平台，支持单聊、群聊、音视频通话，采用 WebSocket 长连接实现实时消息推送，集成 Redis 缓存、Kafka 消息队列、gRPC 远程查询等企业级中间件。

---

## 技术栈总览

| 层级 | 技术选型 | 说明 |
|------|---------|------|
| **后端语言** | Go 1.25 | 高并发、低延迟，原生 goroutine 并发模型 |
| **Web 框架** | Gin v1.10 | 高性能 HTTP 路由，中间件链式处理 |
| **前端框架** | Vue 3 + Element Plus | 组合式 API，响应式 UI |
| **状态管理** | Vuex 4 + Vue Router 4 | 前端状态集中管理与路由控制 |
| **ORM** | GORM v1.25 | 自动迁移、软删除、事务管理 |
| **数据库** | MySQL | 关系型存储 |
| **缓存** | Redis (go-redis/v8) | Cache-Aside 模式，TTL 过期策略 |
| **消息队列** | Apache Kafka (segmentio/kafka-go) | 异步消息投递，Hash 分区策略 |
| **实时通信** | gorilla/websocket | 全双工 WebSocket 长连接 |
| **音视频** | WebRTC + coturn (TURN) | P2P 音视频通话，WebSocket 信令 |
| **RPC 框架** | gRPC + Protocol Buffers | 查询服务远程调用 |
| **短信服务** | 阿里云 SMS SDK | 手机验证码登录与注册 |
| **日志系统** | Zap + Lumberjack | 结构化 JSON 日志，自动轮转 |
| **安全** | TLS/HTTPS + AES-CFB 加密 | 传输层加密 + 数据加密 |
| **配置管理** | TOML (BurntSushi/toml) | 多模块配置 |
| **部署** | Ubuntu + Apache2 反向代理 + systemd | 生产级部署方案 |

---

## 技术要点与设计亮点

### 1. WebSocket 实时消息引擎

 **goroutine-per-connection** 并发模型，每个 WebSocket 连接维护独立的 Read 和 Write 协程，通过 channel 实现协程间解耦通信。

**Chat Server 事件循环**：核心消息路由基于 `select` 多路复用，同时监听 Login / Logout / Transmit 三个 channel，实现非阻塞事件分发。

**消息路由策略**：
- 接收者 ID 以 `U` 开头 → 单聊，直接投递到目标用户的 `SendTo` channel
- 接收者 ID 以 `G` 开头 → 群聊，解析 Members 列表后 fan-out 广播
- 每条消息同时回显给发送者，保证客户端 UI 一致性

**背压处理**：当目标用户的 `SendTo` channel 满（缓冲区 20）时，降级为 Kafka 异步投递，防止阻塞并避免消息丢失。

### 2. 双模式消息投递（Channel / Kafka）



| 模式 | 实现 | 适用场景 |
|------|------|---------|
| **Channel** | Go 原生 channel，内存级传输 | 单机部署，低延迟（默认） |
| **Kafka** | 消息写入 Kafka topic，消费者异步处理 | 分布式部署，高可靠、消息持久化 |

Kafka 生产者采用 Hash 分区策略，确保同一会话消息有序；消费者使用 Consumer Group 模式，支持水平扩展。

### 3. Redis Cache-Aside 缓存策略

      Cache-Aside 模式：优先读取 Redis 缓存，缓存未命中时查询 MySQL 并回填缓存设置 TTL；写请求更新 MySQL 后主动删除受影响的缓存 Key。


**一致性保障**：写操作后通过前缀/模式匹配批量删除关联缓存 Key，使用 `SCAN` 替代 `KEYS` 命令避免 Redis 线程阻塞。

### 4. 查询服务策略模式（Local / gRPC）

  `Service` 接口（16 个查询方法），运行时根据配置选择实现：

- **Local 模式**：直接调用 GORM 服务层，零网络开销
- **gRPC 模式**：通过 Protobuf 远程调用，支持查询服务独立部署和水平扩展

**gRPC 服务划分**（5 个独立服务）：

| 服务 | 方法数 | 职责 |
|------|--------|------|
| MessageQueryService | 2 | 单聊/群聊消息查询 |
| SessionQueryService | 2 | 用户/群聊会话查询 |
| UserQueryService | 2 | 用户信息 & 列表查询 |
| GroupQueryService | 5 | 群组信息、成员、加群方式查询 |
| ContactQueryService | 5 | 联系人列表、详情、申请列表查询 |

 Protocol Buffers IDL 定义接口，自动生成类型安全的客户端/服务端代码，相比 JSON 序列化减少约 50%-80% 网络传输体积。业务代码无需修改即可通过配置切换调用方式。


### 5. WebRTC 音视频信令

WebSocket 作为信令通道，承载 WebRTC 的 SDP Offer/Answer 和 ICE Candidate 交换。消息类型 `Type=3` 专用于音视频信令，`AVdata` 字段透传 SDP/ICE 序列化数据。部署 coturn TURN 服务器确保 NAT 穿越能力。

### 6. 结构化日志 & 优雅关闭

- **日志**：基于 Zap 构建，双输出（控制台 + 文件），JSON 结构化格式，通过 `runtime.Caller` 记录调用源位置，Lumberjack 自动按大小轮转
- **优雅关闭**：监听 SIGINT/SIGTERM 信号，退出前依次清理 Redis 缓存数据、关闭 Kafka Writer/Reader，避免缓存脏数据和消息队列资源泄漏

---

## 实时消息流转机制

### Channel 模式（单机）

```
 用户A (Writer)                                         用户B (Reader)
     │                                                       ▲
     │ WebSocket                                  WebSocket  │
     ▼                                                       │
 Client.Read()                                     Client.Write()
     │                                                       ▲
     │ SendTo chan                                 SendTo chan│
     ▼                                                       │
 ChatServer.TransmitChan ──► select 事件循环 ──► 路由分发 ───┘
                                    │
                                    ├──► MySQL 持久化 (INSERT message)
                                    ├──► Redis 缓存更新 (DEL + SET)
                                    └──► Session 最新消息更新
```

### Kafka 模式（分布式）

```
 用户A (Writer)                                         用户B (Reader)
     │                                                       ▲
     │ WebSocket                                  WebSocket  │
     ▼                                                       │
 Client.Read()                                     Client.Write()
     │                                                       ▲
     │ Kafka Producer                             SendTo chan│
     ▼                                                       │
 chat_message Topic ──► Kafka Consumer ──► 消息处理 ───────┘
   (Hash 分区)            (Consumer Group)       │
                                                  ├──► MySQL 持久化
                                                  ├──► Redis 缓存更新
                                                  └──► Session 最新消息更新
```
