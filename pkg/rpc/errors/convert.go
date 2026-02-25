package errors

import "encoding/json"

// FromProtocolResponse 将协议 status + body 转为 Go error。
func FromProtocolResponse(statusCode uint16, body []byte) error {
	if statusCode == uint16(OK) {
		return nil
	}

	rpcErr := &RpcError{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, rpcErr); err == nil {
			if rpcErr.Code == 0 {
				rpcErr.Code = Code(statusCode)
			}
			if rpcErr.Message == "" {
				rpcErr.Message = "rpc error"
			}
			return rpcErr
		}
	}

	message := "rpc error"
	if len(body) > 0 {
		message = string(body)
	}
	return &RpcError{
		Code:    Code(statusCode),
		Message: message,
	}
}
