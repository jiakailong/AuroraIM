package codec

import (
	"encoding/json"
	"kama_chat_server/pkg/rpc/protocol"
)

type JSONCodec struct{}

func NewJSONCodec() Codec {
	return JSONCodec{}
}

func (JSONCodec) Name() string {
	return "json"
}

func (JSONCodec) ID() uint8 {
	return protocol.CodecJSONID
}

func (JSONCodec) Marshal(value any) ([]byte, error) {
	return json.Marshal(value)
}

func (JSONCodec) Unmarshal(data []byte, value any) error {
	return json.Unmarshal(data, value)
}
