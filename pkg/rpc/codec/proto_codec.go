package codec

import (
	"fmt"
	"kama_chat_server/pkg/rpc/protocol"
)

// ProtoCodec 为 Protobuf 编解码预留实现。
type ProtoCodec struct{}

func NewProtoCodec() Codec {
	return ProtoCodec{}
}

func (ProtoCodec) Name() string {
	return "protobuf"
}

func (ProtoCodec) ID() uint8 {
	return protocol.CodecProtobufID
}

func (ProtoCodec) Marshal(value any) ([]byte, error) {
	return nil, fmt.Errorf("%w: %s", ErrUnsupportedCodec, "protobuf marshal is not implemented yet")
}

func (ProtoCodec) Unmarshal(data []byte, value any) error {
	return fmt.Errorf("%w: %s", ErrUnsupportedCodec, "protobuf unmarshal is not implemented yet")
}
