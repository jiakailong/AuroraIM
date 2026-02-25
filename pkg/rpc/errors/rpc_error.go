package errors

import "fmt"

// RpcError 是 RPC 统一错误结构。
type RpcError struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (errorInfo *RpcError) Error() string {
	if errorInfo == nil {
		return "rpc error: <nil>"
	}
	if errorInfo.Message == "" {
		return fmt.Sprintf("rpc error code=%d", errorInfo.Code)
	}
	return fmt.Sprintf("rpc error code=%d message=%s", errorInfo.Code, errorInfo.Message)
}

func New(code Code, message string, details any) *RpcError {
	return &RpcError{
		Code:    code,
		Message: message,
		Details: details,
	}
}
