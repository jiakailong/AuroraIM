package api

// RequestEnvelope 为通用 RPC 请求基础类型。
type RequestEnvelope struct {
	RequestID uint64   `json:"request_id"`
	Method    string   `json:"method"`
	Metadata  Metadata `json:"metadata,omitempty"`
	Payload   []byte   `json:"payload,omitempty"`
}

// ResponseEnvelope 为通用 RPC 响应基础类型。
type ResponseEnvelope struct {
	RequestID uint64   `json:"request_id"`
	Code      uint16   `json:"code"`
	Message   string   `json:"message,omitempty"`
	Metadata  Metadata `json:"metadata,omitempty"`
	Payload   []byte   `json:"payload,omitempty"`
}
