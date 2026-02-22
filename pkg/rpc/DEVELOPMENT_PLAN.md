# KamaChat 自研 RPC 框架开发计划

## 1. 目标与范围

### 1.1 总体目标
- 在项目内实现一套自研 RPC 框架，代码统一放在 `pkg/rpc`。
- 先满足 Go-to-Go 内部调用，再逐步接入现有 KamaChat 业务。
- 保持现有前端接口和 WebSocket 协议不变，优先实现平滑迁移。

### 1.2 V1 范围（必须完成）
- TCP 长连接通信。
- 请求-响应模型（非流式）。
- 服务注册与方法调用。
- 超时控制、基础重试、心跳保活。
- 错误码规范、日志与基础指标。
- 消息查询服务接入（先读后写）。

### 1.3 V2 范围（后续迭代）
- 动态服务发现（etcd/nacos）。
- 熔断、限流、降级策略。
- 更完善的追踪与监控。
- 多序列化协议（JSON/Protobuf 可切换）。

---

## 2. 设计原则
- **简单可用优先**：先做可运行 MVP，再增强治理能力。
- **低侵入改造**：接入业务时尽量少改已有 Controller 和 DTO。
- **可扩展协议**：协议头预留版本、压缩、序列化等字段。
- **可观测性内建**：从第一版开始加入日志、trace_id、请求统计。
- **灰度迁移**：每个业务接口支持 local/rpc 双模式切换。

---

## 3. 目标项目结构（推荐）

```text
pkg/
	rpc/
		DEVELOPMENT_PLAN.md
		README.md
		api/
			types.go                 # 通用 Request/Response 定义
			context.go               # 元数据、trace_id、deadline
		protocol/
			header.go                # 协议头定义
			frame.go                 # 编帧/拆帧
			message.go               # 消息结构定义
			constants.go             # 魔数、版本、消息类型
		codec/
			codec.go                 # Codec 接口
			json_codec.go            # JSON 编解码
			proto_codec.go           # 预留：Protobuf 编解码
		transport/
			server.go                # TCP 监听与连接接入
			client_conn.go           # 客户端连接封装
			conn_pool.go             # 连接池
			heartbeat.go             # 心跳与保活
		registry/
			registry.go              # 注册发现接口
			static_registry.go       # 静态注册实现
			memory_registry.go       # 本地内存实现（测试）
		balancer/
			balancer.go              # 负载均衡接口
			random.go                # 随机策略
			round_robin.go           # 轮询策略
			consistent_hash.go       # 一致性哈希（可后置）
		server/
			server.go                # RPC Server 启动与生命周期
			service_register.go      # 服务注册（反射）
			dispatcher.go            # 请求分发
			handler.go               # 方法调用适配
		client/
			client.go                # RPC Client 门面
			invoke.go                # Call/Invoke 实现
			retry.go                 # 重试策略
			options.go               # 超时、重试、序列化等配置
		middleware/
			middleware.go            # 中间件链
			logging.go               # 请求日志
			recovery.go              # panic 恢复
			auth.go                  # 元数据鉴权（可选）
			metrics.go               # 指标采集
		errors/
			code.go                  # 统一错误码
			rpc_error.go             # 错误结构体
			convert.go               # 业务错误映射
		observability/
			logger.go                # 框架日志适配
			metrics.go               # 指标出口（Prometheus 预留）
			tracing.go               # trace_id 透传
		config/
			config.go                # RPC 配置结构
			loader.go                # 配置加载
		examples/
			echo/
				server/main.go
				client/main.go
			message_query/
				server/main.go
				client/main.go
		tests/
			protocol_test.go
			codec_test.go
			client_server_test.go
			timeout_retry_test.go
			benchmark_test.go
```

---

## 4. 与现有项目的接入结构规划

### 4.1 新增命令入口
```text
cmd/
	rpc_message_server/
		main.go
	rpc_user_server/
		main.go
	rpc_session_server/
		main.go
```

### 4.2 业务分层建议
- `internal/service/gorm` 保留为本地业务实现。
- 新增 `internal/service/rpc_client` 作为 RPC 调用封装层。
- Controller 继续调用 service 层，不直接感知 RPC 细节。

### 4.3 首批接入接口（低风险）
- 消息查询：
	- `/message/getMessageList`
	- `/message/getGroupMessageList`
- 会话查询：
	- `/session/getUserSessionList`
	- `/session/getGroupSessionList`

---

## 5. 分阶段开发计划

## 阶段 A：框架骨架（第 1 周）

### 目标
- 搭建目录、核心接口、协议定义。

### 任务
- 定义 `Request/Response` 标准结构。
- 定义协议头字段：magic/version/msgType/codec/requestId/timeout/bodyLen。
- 完成编帧拆帧工具。
- 实现 JSON codec。

### 里程碑
- 本地单测通过：协议编解码正确。

---

## 阶段 B：最小可用 RPC（第 2 周）

### 目标
- 跑通 1 个服务端 + 1 个客户端同步调用。

### 任务
- 实现 `server.Register()`（反射注册方法）。
- 实现 `client.Call(ctx, service, method, req, resp)`。
- 处理 requestId 映射，支持并发请求。
- 完成基础错误返回。

### 里程碑
- `examples/echo` 可运行，支持并发 100 请求。

---

## 阶段 C：可用性增强（第 3 周）

### 目标
- 补齐生产必需能力。

### 任务
- 客户端连接池。
- 超时控制与 context 取消。
- 心跳与断连重连。
- 幂等请求的重试策略（默认关闭，按接口开启）。

### 里程碑
- 断开服务端后客户端可恢复。
- 超时与取消行为可预测。

---

## 阶段 D：治理与观测（第 4 周）

### 目标
- 完成基本中间件与可观测能力。

### 任务
- 日志中间件（包含 requestId、service、method、耗时）。
- Recovery 中间件（panic 保护）。
- Metrics 指标（请求总数、错误数、耗时分布）。
- 统一错误码映射规则。

### 里程碑
- 排障时可追踪一次调用链。

---

## 阶段 E：接入消息查询服务（第 5 周）

### 目标
- 将消息查询链路迁移到自研 RPC。

### 任务
- 新增 `cmd/rpc_message_server/main.go`。
- 提供 RPC 方法：
	- `MessageService.GetMessageList`
	- `MessageService.GetGroupMessageList`
- 在 `internal/service/gorm/message_service.go` 增加切换逻辑：
	- `mode=local` 调用原 DAO
	- `mode=rpc` 调用 RPC Client

### 里程碑
- 两个消息查询接口 local/rpc 返回一致。

---

## 阶段 F：接入用户与会话查询（第 6 周）

### 目标
- 扩展 RPC 价值面，提升架构一致性。

### 任务
- 新增 `rpc_user_server`、`rpc_session_server`。
- 将读接口优先迁移（不先改写接口）。
- 补齐接口级别超时与重试配置。

### 里程碑
- 核心查询类接口全部可走 RPC。

---

## 阶段 G：消息写链路评估与灰度（第 7-8 周）

### 目标
- 评估并推进写链路 RPC 化（可选，需灰度）。

### 任务
- 对 `chat/server.go`、`chat/kafka_server.go` 的写 DB 部分做抽象。
- 增加 `MessageCommand` RPC 服务（创建消息、更新状态）。
- 用配置开关灰度发布，出现异常可一键回退。

### 里程碑
- 在灰度环境稳定运行，错误率可控。

---

## 6. 开发规范与编码约定
- 框架层禁止依赖业务包（`internal/*`），必须保持独立。
- 所有导出接口写 GoDoc 注释。
- 每个公共方法返回可判定错误码。
- 核心链路必须支持 `context.Context`。
- 单文件不超过 400 行，复杂逻辑拆分子文件。

---

## 7. 测试计划

### 7.1 单元测试
- 协议头序列化与兼容测试。
- 编解码正确性测试。
- requestId 并发映射测试。
- 超时、取消、重试行为测试。

### 7.2 集成测试
- 启动 server/client 后真实 TCP 调用。
- 服务不可用、重连恢复测试。
- 多实例 + 轮询负载均衡测试。

### 7.3 性能测试
- 基线：P50/P95 延迟、QPS、错误率。
- 压测对比 local 调用与 rpc 调用差异。

---

## 8. 配置与灰度策略

### 8.1 配置项建议
- `rpc.enabled`
- `rpc.mode`（local/rpc）
- `rpc.timeout_ms`
- `rpc.retry.max`
- `rpc.registry.type`（static/memory/etcd）
- `rpc.services.<name>.endpoints`

### 8.2 灰度步骤
1. 测试环境全量启用 RPC。
2. 线上先开查询接口。
3. 观察日志和指标稳定后扩大比例。
4. 写链路最后切换，并保留回退开关。

---

## 9. 风险清单与应对
- **风险：协议演进不兼容**
	- 应对：协议头带 version，新增字段向后兼容。
- **风险：连接泄漏/协程泄漏**
	- 应对：统一连接生命周期管理，加入泄漏测试。
- **风险：重试导致重复写**
	- 应对：写接口默认不重试，幂等接口才重试。
- **风险：迁移期间行为不一致**
	- 应对：双模式对比日志，按接口灰度。

---

## 10. 交付物清单
- `pkg/rpc` 完整框架代码。
- `pkg/rpc/README.md` 使用文档。
- `pkg/rpc/examples` 可运行示例。
- 接入层代码（message/user/session）。
- 测试与压测报告。
- 灰度发布与回退手册。

---

## 11. 启动动作（今天就能做）
1. 创建 `pkg/rpc` 目录骨架。
2. 先实现 `protocol + codec + server/client 最小调用`。
3. 写 `examples/echo` 跑通第一条 RPC。
4. 同步接入 `MessageService.GetMessageList` 作为第一条业务链路。

> 建议节奏：先跑通，再治理，再扩展。避免一开始做成大而全，导致接入延迟。

---

## 12. 每周开发计划（按项目结构落地）

> 执行规则：每周开始前先阅读 `pkg/rpc/DEVELOPMENT_SKILLS.md`，完成“开发前检查清单”后再开始编码。

### Week 1：框架骨架与协议定稿

**目标**
- 完成 `api/protocol/codec` 三层骨架。
- 定义并冻结 V1 协议头字段。

**开发内容（对应目录）**
- `pkg/rpc/api`：Request/Response、Meta、Context。
- `pkg/rpc/protocol`：Header、Frame、消息常量、编码边界。
- `pkg/rpc/codec`：Codec 接口 + JSON 实现。

**交付物**
- 可执行的协议编解码单测。
- 协议文档（字段定义+兼容规则）。

**验收标准（DoD）**
- 协议 encode/decode 1000 次随机输入无 panic。
- 错误输入可稳定返回统一错误码。

### Week 2：Server/Client MVP 打通

**目标**
- 实现最小可用 RPC 调用链（同步请求响应）。

**开发内容（对应目录）**
- `pkg/rpc/server`：服务注册、反射路由、请求分发。
- `pkg/rpc/client`：Call/Invoke、requestId 映射。
- `pkg/rpc/transport`：TCP 连接建立、读写循环。

**交付物**
- `examples/echo` 双端示例可运行。
- 并发请求匹配测试。

**验收标准（DoD）**
- 100 并发调用下响应无错配。
- 客户端与服务端异常均可定位日志。

### Week 3：可用性增强（超时/重试/心跳）

**目标**
- 补齐生产必需能力，降低网络抖动影响。

**开发内容（对应目录）**
- `pkg/rpc/client/retry.go`：幂等重试策略。
- `pkg/rpc/transport/heartbeat.go`：心跳与断连重连。
- `pkg/rpc/transport/conn_pool.go`：连接复用。

**交付物**
- 超时取消测试、断连恢复测试。

**验收标准（DoD）**
- 服务端短暂重启后客户端可恢复调用。
- 超时请求能在预期窗口内释放资源。

### Week 4：治理能力（中间件/错误码/可观测）

**目标**
- 建立统一治理底座，支撑后续业务接入。

**开发内容（对应目录）**
- `pkg/rpc/middleware`：logging/recovery/metrics。
- `pkg/rpc/errors`：统一错误码与错误映射。
- `pkg/rpc/observability`：trace_id、耗时统计。

**交付物**
- 调用日志规范 + 指标定义文档。

**验收标准（DoD）**
- 每次 RPC 调用可追踪 requestId、service、method、耗时。
- panic 可被 recovery 拦截并返回标准错误。

### Week 5：接入消息查询服务（第一条业务链路）

**目标**
- 将消息查询接口切换为可选 RPC 模式。

**开发内容（对应目录）**
- `cmd/rpc_message_server/main.go`：消息查询 RPC 服务进程。
- `internal/service/rpc_client`（新增）：消息查询客户端封装。
- `internal/service/gorm/message_service.go`：`local/rpc` 双模式。

**交付物**
- 两个消息查询接口在 local/rpc 模式返回一致性报告。

**验收标准（DoD）**
- `GetMessageList`、`GetGroupMessageList` 可配置切换。
- 切换失败可快速回退 local。

### Week 6：接入用户与会话查询

**目标**
- 扩展查询类场景，形成统一接入模板。

**开发内容（对应目录）**
- `cmd/rpc_user_server/main.go`。
- `cmd/rpc_session_server/main.go`。
- `internal/service/rpc_client`：user/session 调用封装。

**交付物**
- 用户与会话查询接口接入说明文档。

**验收标准（DoD）**
- 查询接口可稳定运行在 RPC 模式。
- 错误率与延迟满足基线目标。

### Week 7：写链路抽象与灰度准备

**目标**
- 为消息写入链路 RPC 化做抽象，不直接全量切换。

**开发内容（对应目录）**
- `internal/service/chat`：抽离消息持久化调用点。
- `pkg/rpc/examples/message_query` 扩展为 command/query 分离示例。
- 增加写链路灰度开关。

**交付物**
- 写链路改造设计评审文档。

**验收标准（DoD）**
- 写链路已具备可切换能力，但默认仍走旧路径。

### Week 8：写链路小流量灰度

**目标**
- 小范围验证消息写入 RPC 化的稳定性。

**开发内容（对应目录）**
- `cmd/rpc_message_server` 增加 `MessageCommand`。
- 配置中心/配置文件接入灰度比例参数。
- 线上观察日志、错误率、消息一致性。

**交付物**
- 灰度报告（成功率、P95、回退记录）。

**验收标准（DoD）**
- 灰度期间无严重消息丢失或重复写事故。
- 支持一键回退至 local 逻辑。

### Week 9：性能优化与稳定性加固

**目标**
- 在功能稳定后优化吞吐与延迟。

**开发内容（对应目录）**
- 连接池参数调优、序列化优化。
- 热点接口压测与瓶颈分析。
- 引入更细粒度性能指标。

**交付物**
- 性能压测报告 + 参数建议。

**验收标准（DoD）**
- 相比 Week 6 版本，P95 或 QPS 有明确提升。

### Week 10：发布固化与文档交付

**目标**
- 完成框架沉淀与团队交接。

**开发内容（对应目录）**
- `pkg/rpc/README.md` 完整化。
- 运维发布手册、排障手册、回退手册。
- 统一代码模板和新服务接入模板。

**交付物**
- 可复用 RPC 框架 v1.0。
- 团队可执行的开发与运维手册。

**验收标准（DoD）**
- 新业务可按模板在 1 天内完成 RPC 接入。

---

## 13. 开发前必读机制（Skills Gate）

### 13.1 机制定义
- 每次开始开发前，必须阅读 `pkg/rpc/DEVELOPMENT_SKILLS.md`。
- 未完成文档中的“开发前检查清单”，不得开始编码。

### 13.2 执行方式
1. 在任务开始时勾选 skills 清单。
2. 提交 PR 时附上“skills 清单截图或勾选记录”。
3. Code Review 先审规范合规，再审业务逻辑。

### 13.3 审核重点
- 是否违反框架层依赖边界。
- 是否遗漏超时、错误码、日志与 context 透传。
- 是否补充测试与回归验证。
