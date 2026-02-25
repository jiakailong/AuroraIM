# KamaChat 技术报告：项目全景、RPC集成现状与后续规划（2026-02-25）

## 1. 项目整体介绍

### 1.1 聊天项目（KamaChat）架构概览
KamaChat 当前是一个以 `Gin + Gorm + WebSocket + Redis + Kafka` 为核心的聊天系统，核心结构为：
- 接入层：`api/v1/*_controller.go`（HTTP/WS 接口入口）
- 业务层：`internal/service/gorm`（主要业务实现） + `internal/service/chat`（WebSocket收发）
- 数据层：`internal/dao` + `internal/model`
- 基础设施：Redis、Kafka、MySQL
- 启动入口：`cmd/kama_chat_server/main.go`


`Controller -> gorm service -> dao/model -> DB/缓存/消息队列`。

### 1.2 RPC 框架（pkg/rpc）能力概览
.git .gitignore .idea .vscode LICENSE README.md api cmd configs deploy_kamachat.sh docs go.mod go.sum go1.20.linux-amd64.tar.gz internal pkg start_service.sh web  test RPC 框架位于 `pkg/rpc`，已完成：
- 自定义二进制协议（fixed header + var header + body）
- 编解码层（JSON + Proto占位）
- TCP 长连接、连接池、心跳保活
- 服务注册与反射调用
- 服务发现（memory/static）、负载均衡（RR/Random）、failover
- 幂等约束下重试（Unavailable/Timeout）
- middleware（tracing / metrics / logging / recovery）
- examples（echo、message_query）与完整测试集

RPC调用路径（当前实现）为：
`Client -> registry/balancer -> conn pool -> transport -> server dispatcher -> middleware chain -> reflected handler`。

---

## 2. 当前哪些服务在使用 RPC 调用

## 2.1 生产主进程是否支持 RPC 查询链路

- 配置位于 `queryConfig`：`mode = local | rpc`
- 初始化位置：`cmd/kama_chat_server/main.go` -> `query.Init(...)`
- 统一门面：`internal/service/query`

ls**同一套 Controller，可在启动时切换本地查询或 RPC 查询**。

## 2.2 已接入可切换（local/rpc）的查询服务
 `query.QueryService` 已接入：

1) 消息查询
- `GetMessageList`
- `GetGroupMessageList`

2) 会话查询
- `GetUserSessionList`
- `GetGroupSessionList`

3) 用户查询
- `GetUserInfoList`
- `GetUserInfo`

4) 群查询
- `LoadMyGroup`
- `CheckGroupAddMode`
- `GetGroupInfo`
- `GetGroupInfoList`
- `GetGroupMemberList`

5) 联系人查询
- `GetUserList`
- `LoadMyJoinedGroup`
- `GetContactInfo`
- `GetNewContactList`
- `GetAddGroupList`

## 2.3 尚未切换为 RPC 的部分
- 写链路（如注册、登录、建群、删除、更新、通过申请、拉黑等）仍保持本地调用。
- 这是有意设计：先完成低风险查询链路，再逐步推进写链路。

---

## 3. 性能测试与“提升了多少”

## 3.1 测试方法（自行实测）
ls
- `go test -run '^$' -bench 'Benchmark(Local|RPC)_Echo|BenchmarkRPC_NoPool_Echo' -benchmem ./pkg/rpc/tests`

ls
- `BenchmarkLocal_Echo`：本地函数直调（基线）
- `BenchmarkRPC_Echo`：RPC + 连接池（当前推荐）
- `BenchmarkRPC_NoPool_Echo`：RPC 无连接复用（退化场景）

## 3.2 实测结果
- `BenchmarkLocal_Echo-16`: `9.192 ns/op`, `16 B/op`, `1 allocs/op`
- `BenchmarkRPC_Echo-16`: `17306 ns/op`, `5365 B/op`, `96 allocs/op`
- `BenchmarkRPC_NoPool_Echo-16`: `368682 ns/op`, `33266 B/op`, `157 allocs/op`

## 3.3 结果解读（重点）
1) 相比本地直调，RPC 不会更快（单次调用有序列化与网络开销）
- 延迟比：$\frac{17306}{9.192} \approx 1883$ 倍（RPC更慢）
- 结论：RPC 的价值不是单次调用提速，而是**解耦与可扩展性**。

2) 在 RPC 框架内部，连接池优化带来显著提升
- 延迟下降：$\frac{368682 - 17306}{368682} \approx 95.31\%$
- 吞吐提升（近似）：$\frac{368682}{17306} \approx 21.30$ 倍
- 结论：相比“每次新建连接”，当前连接池方案已显著优化性能。

## 3.4 对“性能提升了多少”的结论
- 若对比“本地直调”：没有提升，RPC有固定开销。
- 若对比“低效RPC（无连接池）”：当前RPC实现在延迟上提升约 **95.31%**，吞吐约提升 **21.30x**。
- 若对比“历史线上链路”：需补同口径 A/B 压测后给出最终业务提升值。

---

## 4. 后续改进建议

## 4.1 RPC 框架层改进
1) 注册中心升级
- 从 static/memory 升级到 etcd/nacos（动态发现 + 健康检查）。

2) 治理能力增强
- 熔断、限流、隔离仓、超时预算、重试预算（按方法级配置）。

3) 指标体系落地
- 将 in-memory metrics 对接 Prometheus，补齐 p95/p99、错误码分布、实例维度。

4) 协议与序列化优化
- 完成 Protobuf 正式实现，评估压缩与零拷贝优化。

5) 稳定性保障
- 增加故障注入测试（丢包、抖动、半开连接、慢节点）。

## 4.2 业务侧 RPC 化路线建议
#
#-
ls
1. 查询链路（已完成主体）继续做灰度和观测
2. 读写弱一致链路
3. 核心写链路（消息发送、申请处理）

ls
- 幂等键（如 `clientMsgID`）
- 去重机制
- 重试安全策略
- 补偿与回滚机制

## 4.3 推荐下一批升级服务
- 会话写接口（较易控制）
- 群成员变更写接口（中等复杂）
- 消息发送主写链路（最后推进）

---

## 5. 运行与配置说明（当前已支持）
pkg/rpc/tests/benchmark_test.go--------的 `queryConfig`：
- `mode = "local"`：查询走本地 gorm
- `mode = "rpc"`：查询走 RPC client
- `rpcListen`：内置查询 RPC server 监听地址（可选）
- `rpcTarget`：查询 RPC 目标地址（可选，默认回落监听地址）

#ls
#ls


---

## 6. 报告边界说明
#
ls --color=auto --color=auto --color=auto =auto
