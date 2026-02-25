# message_query example

## 目标
ls Day14 最小聊天闭环：
- A 通过网关发送消息（RPC `SendMessage`）
- MessageService 返回 ack
- B 拉取离线消息（RPC `PullOffline`）

## 运行方式
1. 启动 MessageService RPC 服务：
   - go run ./pkg/rpc/examples/message_query/server -addr 127.0.0.1:19100
2. 方式 A（自动演示）：
   - go run ./pkg/rpc/examples/message_query/client
3. 方式 B（网关模拟）：
   - go run ./pkg/rpc/examples/message_query/gateway -action send -from userA -to userB -content "hello" -client_msg_id msg-1
   - go run ./pkg/rpc/examples/message_query/gateway -action pull -user userB

## 说明
- 存储为内存实现，仅用于示例验证。
- `client_msg_id` 用于去重，同一发送者重复发送同一个 `client_msg_id` 会返回重复 ack。
