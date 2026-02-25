package codec

import "errors"

var (
	ErrUnsupportedCodec = errors.New("codec: unsupported codec")
)

// Codec 定义 RPC Body 编解码行为。
type Codec interface {
	Name() string
	ID() uint8
	Marshal(value any) ([]byte, error)
	Unmarshal(data []byte, value any) error
}
