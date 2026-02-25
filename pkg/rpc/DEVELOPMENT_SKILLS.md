# DEVELOPMENT_SKILL.md (For Codex)

你正在为 chat/pkg/rpc 开发一个可用的 RPC 框架。每次开始编码前必须阅读本文件与 protocol/PROTOCOL.md。
目标：代码规范、可扩展、可测试、可集成到聊天系统。

---

## 1. 模块边界（必须遵守，避免循环依赖）
- api：只做 context/metadata/types，不依赖 client/server/transport
- protocol：只做二进制协议与 Frame 编解码，不依赖 codec/client/server
- codec：只做序列化实现，不依赖 client/server/transport
- transport：只做连接管理、read/write loop、pool/heartbeat，不理解 “service/method”
- server：负责 Register/dispatch/middleware/handler，调用 protocol/codec/transport
- client：负责 Invoke/Call/pending/timeout/retry/lb/registry/pool，调用 protocol/codec/transport
- middleware：只依赖 api/errors/observability，不直接依赖 transport
- observability：只提供 logger/metrics/tracing 适配接口，不依赖 client/server 具体实现
- errors：统一错误码与 RpcError，任何层都可依赖

若出现循环依赖，优先：提取接口/类型到 api 或 errors 或 observability。

---

## 2. 协议一致性（必须）
- 所有 Frame 必须符合 protocol/PROTOCOL.md 的字段与字节序（BigEndian）。
- 禁止“临时改协议字段”。协议变更必须：
  1) 更新 protocol/PROTOCOL.md
  2) bump Version 或保持向后兼容
  3) 更新 tests/protocol_test.go

---

## 3. 并发与网络 I/O 规则（必须）
- 严禁多个 goroutine 并发写同一个 net.Conn。
  - 必须通过 writeLoop + writeChan 串行写。
- readLoop 只负责解帧与投递上层，不做耗时业务。
- pending map 必须加锁（mutex）或使用 sync.Map；超时/取消必须清理 pending，避免泄漏。
- Close() 必须幂等（sync.Once 或原子标记），并确保 read/write goroutine 可退出。

---

## 4. 超时与取消（必须）
- client.Call/Invoke 必须接受 context.Context。
- ctx 超时/取消时：
  - 立即返回
  - 清理 pending
  - 不允许 goroutine 持续阻塞等待响应
- server 侧执行 handler 使用 ctx（可基于 header TimeoutMs/metadata deadline 取 min）。

---

## 5. 错误语义（必须）
- 统一使用 errors 包的 Code 与 RpcError。
- Response 的 statusCode != OK：
  - client 返回 *errors.RpcError（保留 Code/Message）
- 方法不存在 -> NotFound
- 连接不可用 -> Unavailable
- ctx 超时 -> Timeout
- panic -> Internal（并 recover）

禁止在业务层用 fmt.Errorf("xxx") 直接吞掉错误码。

---

## 6. Middleware（拦截器）规则
- middleware 必须是纯函数式链式：
  - logging、recovery、metrics、auth（可选）
- logging 输出必须包含：
  - method、requestID、trace_id、latency、code
- recovery 必须捕获 panic 并转为 Internal 错误码

---

## 7. 可观测性规则
- 不允许在核心路径直接 fmt.Printf。
- 统一通过 observability/logger.go 的 Logger 接口输出。
- trace_id 必须通过 api/context.go 的 Metadata 透传（键名固定：trace_id）。

---

## 8. 测试与验收（每次提交前必做）
- gofmt：所有文件必须 gofmt
- go test ./... 必须通过
- 涉及并发/网络变更时，至少跑：go test -race ./pkg/rpc/...
- 新增协议/拆包逻辑必须补 tests/protocol_test.go
- 新增 client/server 功能必须补 tests/client_server_test.go 或 timeout_retry_test.go

---

## 9. 代码风格（必须）
- 公开 API（client/server/options）写 GoDoc 注释
- 重要结构体字段写注释（例如 pending/connPool）
- 选项使用 Option Pattern（func(*Options)）
- 避免全局变量；常量放 protocol/constants.go 或 errors/code.go
- 返回错误时尽量 wrap：fmt.Errorf("xxx: %w", err)

---

## 10. 输出要求
每次生成代码必须同时输出：
1) 修改/新增了哪些文件
2) 新增了哪些 public API
3) 对应新增/更新了哪些 tests
4) 如协议/行为变更，指出是否需要更新 PROTOCOL.md