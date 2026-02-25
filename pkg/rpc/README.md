# KamaChat RPC Framework

## Overview
`pkg/rpc` 是 KamaChat 的自研 RPC 框架实现，当前已完成：
- 自定义二进制协议（固定头 + 可变头 + body）
- 编解码（JSON + Protobuf 占位）
- TCP 长连接、心跳保活、连接池
- 服务注册与反射调用
- Registry + Balancer + Failover
- 幂等约束下的重试（Unavailable/Timeout）
- 中间件链（tracing、metrics、logging、recovery）

## Architecture
ls
1. Client 根据 service/registry/balancer 选择实例
2. 从该实例连接池获取连接并发送请求帧
3. Server dispatcher 解帧 -> middleware 链 -> handler
4. 返回响应帧，client pending map 按 request_id 匹配

ls
- `protocol/`: 协议与编解帧
- `transport/`: 连接与服务接入
- `server/`: 分发与反射注册
- `client/`: invoke、failover、retry
- `middleware/`: tracing/logging/metrics/recovery
- `observability/`: logger/metrics/tracing
- `examples/`: `echo` 与 `message_query`

## Quick Start
1. 运行 echo 示例：
   - Server: `go run ./pkg/rpc/examples/echo/server`
   - Client: `go run ./pkg/rpc/examples/echo/client`
2. 运行 message_query 示例：
   - Server: `go run ./pkg/rpc/examples/message_query/server -addr 127.0.0.1:19100`
   - Client: `go run ./pkg/rpc/examples/message_query/client`

## Benchmark
ls
- `go test -bench BenchmarkRPC_Echo -benchmem ./pkg/rpc/tests`


- 吞吐：约 `10^4 ~ 10^5 ops/s`
- 平均延迟：亚毫秒到毫秒级

## Validation
ls
- `go test ./pkg/rpc/...`
- `go test -race ./pkg/rpc/...`
