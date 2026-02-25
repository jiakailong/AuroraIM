# PROTOCOL.md

本协议用于 chat/pkg/rpc 的 TCP RPC 通信。
目标：
- 高性能：二进制头 + 长度字段拆包
- 可扩展：Version/Flags/CodecID/可选 Metadata
- 可治理：RequestID 并发匹配、Timeout、Heartbeat
- 易实现：固定头 + 变长头 + body

所有整数均为 **BigEndian（网络字节序）**。

---

## 1. Frame 总体结构

Frame = FixedHeader (32 bytes) + VarHeader (HeaderLen bytes) + Body (BodyLen bytes)

- FixedHeader：固定 32 字节，快速校验/拆包
- VarHeader：method + metadata（可为空）
- Body：序列化后的 payload（JSON/Protobuf），可选压缩

---

## 2. FixedHeader 字节布局（32 bytes）

| 字段 | 类型 | 字节 | 说明 |
|------|------|------|------|
| Magic | uint16 | 2 | 魔数，用于快速过滤非法包 |
| Version | uint8 | 1 | 协议版本，当前 = 1 |
| MsgType | uint8 | 1 | 消息类型：Request/Response/Ping/Pong |
| Flags | uint16 | 2 | 标志位（压缩/oneway 等） |
| CodecID | uint8 | 1 | 0=default, 1=JSON, 2=Protobuf(预留) |
| Reserved1 | uint8 | 1 | 预留，置 0 |
| Status | uint16 | 2 | 响应状态码：0=OK；请求置 0 |
| Reserved2 | uint16 | 2 | 预留，置 0 |
| HeaderLen | uint32 | 4 | VarHeader 长度（bytes） |
| BodyLen | uint32 | 4 | Body 长度（bytes，若压缩则为压缩后长度） |
| RequestID | uint64 | 8 | 请求唯一 ID，用于并发匹配 |
| TimeoutMs | uint32 | 4 | 客户端期望超时（ms）。0 表示不传，由 ctx/默认值决定 |

### 约束
- HeaderLen <= 64KB（可配置）
- BodyLen <= 16MB（可配置）
- 超限：应直接关闭连接（防止内存攻击）

---

## 3. MsgType 定义

| MsgType | 值 | 说明 |
|--------|----|------|
| Request | 1 | RPC 请求 |
| Response | 2 | RPC 响应 |
| Ping | 3 | 心跳请求 |
| Pong | 4 | 心跳响应 |

---

## 4. Flags 定义（uint16 位图）

| bit | 值 | 含义 |
|-----|----|------|
| 0 | 0x0001 | Body 使用 gzip 压缩 |
| 1 | 0x0002 | OneWay（只发不等响应）预留 |
| 2..15 | - | 预留 |

备注：压缩只作用于 Body；Header 不压缩。

---

## 5. VarHeader 结构（HeaderLen bytes）

VarHeader 用于携带：
- Method（请求必须）
- Metadata（trace_id、auth、deadline 等）

编码格式（顺序固定，便于解析）：

### 5.1 MethodBlock
- MethodLen: uint16
- Method: []byte (UTF-8)

约束：
- Request：MethodLen > 0，Method 格式建议为 `Service.Method`
- Response：MethodLen 可以为 0（可选携带）

### 5.2 MetadataBlock
- MetaCount: uint16
- 循环 MetaCount 次：
  - KeyLen: uint16
  - Key: []byte (UTF-8)
  - ValLen: uint16
  - Val: []byte (UTF-8)

推荐标准键：
- trace_id：链路追踪 id
- deadline_ms：unix 毫秒时间戳（可选，TimeoutMs 也可表达）
- auth_token：鉴权 token（如使用）
- client_addr/server_addr：可选调试字段（避免写敏感信息）

---

## 6. Body 语义

### 6.1 Request Body
- 用 CodecID 指定序列化方式
- Body = Marshal(reqPayload)

### 6.2 Response Body
- 若 Status == 0 (OK)：Body = Marshal(respPayload)
- 若 Status != 0：Body = Marshal(errors.RpcError{Code,Message,Details})
  - client 收到非 OK 状态时，应按 RpcError 解码并返回 error

---

## 7. 心跳（Ping/Pong）

- Ping：MsgType=3，BodyLen=0，RequestID 可为 0
- Pong：MsgType=4，BodyLen=0
- 建议：
  - client 每 N 秒发送 Ping（N 可配置）
  - 超过 M 秒未收到 Pong 或读写失败，判定连接不可用，连接池剔除并重连

---

## 8. 超时语义（TimeoutMs 与 ctx 的关系）

- client 侧：
  - TimeoutMs 可从 ctx deadline 或 options 默认超时推导
- server 侧：
  - 若收到 TimeoutMs > 0，可用于设置 handler 执行的 ctx 超时（取 min(TimeoutMs, ctx deadline)）

---

## 9. 兼容性与版本演进

- FixedHeader 长度固定为 32 bytes（Version=1）。
- 协议新增字段：必须通过 Version 升级或占用 Reserved 字段。
- 修改字段语义：必须升级 Version。
- 若收到未知 Version：
  - 默认拒绝（关闭连接）或返回 Unavailable（由实现决定，但要一致）

---

## 10. 安全建议
- 生产环境建议在 TCP 上启用 TLS（框架可后置实现）。
- Metadata 不应携带敏感明文（如必须带 token，应配合 TLS）。