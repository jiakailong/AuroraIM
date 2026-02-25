package errors

// Code 表示 RPC 统一错误码。
type Code uint16

const (
	OK          Code = 0
	BadRequest  Code = 400
	NotFound    Code = 404
	Timeout     Code = 408
	Internal    Code = 500
	Unavailable Code = 503
)

func (code Code) Uint16() uint16 {
	return uint16(code)
}

func IsSuccess(code Code) bool {
	return code == OK
}
