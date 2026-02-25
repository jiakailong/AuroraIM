#AChat RPC 框架开发计划

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

## Day 1：定协议（第一版）+ Frame 编解码单测
实现：
- protocol/constants.go：Magic/Version/MsgType/Flags/CodecID
- protocol/header.go：FixedHeader 编码/解码（BigEndian），字段校验
- protocol/frame.go：ReadFrame/WriteFrame（io.ReadFull + 长度字段）
- protocol/message.go：Frame/Request/Response 结构定义

测试：
- tests/protocol_test.go：
  - Encode->Decode 一致性
  - 粘包：连续两帧拼在一起可正确拆出
  - 半包：分段 reader 仍能正确解帧
文档：
- protocol/PROTOCOL.md 初稿（先把 header 字节布局写清）

验收：
- `go test ./pkg/rpc/tests -run TestProtocol -v`

---

## Day 2：Codec（先 JSON）+ Body 编解码单测
实现：
- codec/codec.go：Codec 接口（Name/ID/Marshal/Unmarshal）
- codec/json_codec.go：JSON 实现
- codec/proto_codec.go：保留接口与 TODO（不实现也行，但必须能编译）

测试：
- tests/codec_test.go：JSON codec 对 request/response payload 的正确性

验收：
- `go test ./pkg/rpc/tests -run TestCodec -v`

---

## Day 3：Transport Conn（读写循环 + 写队列）+ Echo example
实现：
- transport/client_conn.go：
  - Conn 结构：net.Conn + readLoop + writeLoop + writeChan
  - 只允许 writeLoop 写 socket（避免并发写）
  - Close 语义（幂等关闭、退出循环、清理 chan）
- transport/server.go：
  - Listen/Accept（只做接入，不做 RPC dispatch）
  - 每连接交给上层回调（onFrame / onClose）

示例：
- examples/echo/server/main.go：收到 frame 原样返回
- examples/echo/client/main.go：发送 frame 并打印返回

验收：
- 手动跑 echo：client 能收到一致 frame

---

## Day 4：错误码体系 + api/context 元数据模型
实现：
- errors/code.go：OK/BadRequest/NotFound/Internal/Timeout/Unavailable
- errors/rpc_error.go：RpcError{Code,Message,Details}
- errors/convert.go：将协议 status + body 转为 Go error
- api/context.go：Metadata(trace_id, deadline, auth_token 等) 的 Set/Get（统一 key）
- api/types.go：通用 request/response 基础类型（如 Envelope 可选）

测试：
- 新增单测：RpcError 的编码/解码约定（可复用 JSON）

验收：
- `go test ./pkg/rpc/tests -run TestError -v`（可并入 codec_test）

---

## Day 5：Server MVP（无反射先跑通）+ client/server 集成测试雏形
实现：
- server/server.go：生命周期、Serve(listener)、Close()
- server/dispatcher.go：收到 Request frame -> 找 handler -> 执行 -> 写 Response
- server/handler.go：handler 接口（先手工注册 map[method]func）

测试：
- tests/client_server_test.go：起 server + client 调用成功（先不反射）
验收：
- client 能 Call 到 server 并拿到 response

---

## Day 6：反射服务注册（核心）+ panic recover
实现：
- server/service_register.go：
  - Register(service any)：扫描导出方法
  - 约束签名：func(ctx context.Context, req *T) (*U, error)
  - 生成 methodKey = "Service.Method"
- server/handler.go：反射调用适配（decode req -> call -> encode resp）
- middleware/recovery.go：panic recover（server handler 层）

测试：
- tests/client_server_test.go：使用反射注册的 HelloService 调用成功
- handler panic -> 返回 Internal

验收：
- 方法不存在 -> NotFound
- panic 不崩进程

---

## Day 7：Client Invoke/Call（pending 并发匹配 + timeout）+ 超时单测
实现：
- client/options.go：默认超时/codec/registry/balancer 配置
- client/invoke.go：
  - requestID 生成
  - pending map[requestID]chan resp（mutex保护）
  - ctx 超时/取消时清理 pending
- client/client.go：门面 NewClient/Call/Close

测试：
- tests/timeout_retry_test.go：
  - server sleep > timeout -> client 返回 Timeout
  - pending 清理（至少保证不会永久阻塞）

验收：
- 并发 200+ 调用不串包（建议加一个并发用例）

---

## Day 8：middleware 链（logging + recovery）+ tracing 透传（最小）
实现：
- middleware/middleware.go：Unary middleware chain
- middleware/logging.go：method/requestID/trace_id/latency/code
- observability/logger.go：logger 接口 + 默认实现
- observability/tracing.go：trace_id 生成/透传（写入 metadata）
- api/context.go：确保 trace_id 可取

测试：
- 在 examples/echo 或 message_query 输出日志验证链路
验收：
- 每次调用日志包含 trace_id、耗时、code

---

## Day 9：心跳与保活（ping/pong）+ 连接健康检测
实现：
- transport/heartbeat.go：
  - MsgTypePing/MsgTypePong
  - client 定时 ping；server 回复 pong
  - 超时判定连接不可用（供 pool 使用）

测试：
- 基础集成测试（可写在 client_server_test 或单独 test）
验收：
- server kill 后 client 不会永久卡住（返回 Unavailable/Timeout）

---

## Day 10：连接池（Pool）+ Dial 管理
实现：
- transport/conn_pool.go：
  - Get/Put/Discard
  - maxConn/idleConn
  - 坏连接剔除
- transport/client_conn.go：暴露健康状态/最后活跃时间（供 pool）
- client/client.go/invoke.go：使用 pool 获取连接

测试：
- tests/benchmark_test.go 或单测：并发调用时连接数受控（可打印或用计数器）
验收：
- 不会每次 call 都新建连接

---

## Day 11：Registry（静态/内存）+ 客户端选择实例
实现：
- registry/registry.go：List(service) []Instance（Watch 可先不做）
- registry/static_registry.go：静态配置
- registry/memory_registry.go：测试用
- client/options.go：支持设置 registry 与 serviceName->instances

验收：
- message_query 支持配置多个 server 地址

---

## Day 12：负载均衡（RR/Random）+ failover
实现：
- balancer/balancer.go：Pick(instances, key) -> Instance
- balancer/round_robin.go：轮询
- balancer/random.go：随机（可选）
- client/invoke.go：
  - registry list -> balancer pick -> call
  - 失败时 failover（换下一个实例，最多 1-2 次）

验收：
- 起 2 个 server：调用分摊
- 干掉一个：还能自动换另一个

---

## Day 13：Retry（谨慎）+ 幂等约束（为聊天落地做铺垫）
实现：
- client/retry.go：退避策略（exponential backoff + jitter）
- client/invoke.go：只对可重试 code（Unavailable/Timeout）触发
- 强约束：需要 option 标记该方法/请求为 Idempotent 才允许重试

（聊天侧准备）
- message_query 的 SendMessage 增加 clientMsgID（为去重）

测试：
- tests/timeout_retry_test.go：模拟 Unavailable -> 重试成功（幂等场景）
验收：
- 非幂等默认不重试

---

## Day 14：集成到聊天（先 examples/message_query，再接真实 Gateway/Message）
实现：
- examples/message_query：
  - server：MessageService.SendMessage / PullOffline（可简化存储）
  - client：调用并展示结果
- 聊天集成（最小闭环）：
  - Gateway：收到 WS send -> rpc client Call SendMessage
  - MessageService：RPC Server，落库/写缓存（或先内存+日志）

验收：
- A 发消息 -> Gateway RPC -> MessageService -> 返回 ack -> WS 返回
- B 离线 -> PullOffline 拉取（建议实现）

---

## Day 15：压测/整理/文档收尾（让它“面试可讲”）
实现：
- tests/benchmark_test.go：并发压测（打印 qps/avg/p95）
- middleware/metrics.go + observability/metrics.go：最简指标（计数/耗时/inflight），Prom 预留接口即可
- README.md：架构图、运行方式、协议链接、bench 结果
- DEVELOPMENT_PLAN.md 更新完成情况

验收：
- `go test ./...`
- `go test -race ./pkg/rpc/...`（至少 rpc 相关包）
- examples 一键可运行

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

## 12. 完成状态（截至 2026-02-25）
- Day1 ~ Day15 已完成最小可用实现与测试闭环。
- `pkg/rpc/tests/benchmark_test.go` 已提供并发基准测试入口。
- 已接入 middleware metrics 与 observability metrics（内存指标聚合，Prometheus 可在后续适配）。
- `pkg/rpc/README.md` 已补充架构、运行方式、校验命令。
- 验证结果：`go test ./pkg/rpc/...` 与 `CGO_ENABLED=1 go test -race ./pkg/rpc/...` 均通过。
