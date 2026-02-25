package protocol

const (
	FixedHeaderSize = 32
	CurrentVersion  = uint8(1)
)

const (
	CodecUnknownID  = uint8(0)
	CodecJSONID     = uint8(1)
	CodecProtobufID = uint8(2)

	DefaultCodecID = CodecJSONID
)

const (
	MagicNumber = uint16(0xCAFE)
)

const (
	MsgTypeRequest  = uint8(1)
	MsgTypeResponse = uint8(2)
	MsgTypePing     = uint8(3)
	MsgTypePong     = uint8(4)
)

const (
	FlagGzip   = uint16(0x0001)
	FlagOneWay = uint16(0x0002)
)

const (
	StatusOK          = uint16(0)
	StatusBadRequest  = uint16(400)
	StatusNotFound    = uint16(404)
	StatusTimeout     = uint16(408)
	StatusInternal    = uint16(500)
	StatusUnavailable = uint16(503)
)

const (
	MaxHeaderLen = 64 * 1024
	MaxBodyLen   = 16 * 1024 * 1024
)
