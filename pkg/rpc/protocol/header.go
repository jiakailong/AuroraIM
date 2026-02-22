package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var (
	ErrInvalidMagic     = errors.New("protocol: invalid magic")
	ErrUnsupportedVer   = errors.New("protocol: unsupported version")
	ErrInvalidMsgType   = errors.New("protocol: invalid message type")
	ErrHeaderTooLarge   = errors.New("protocol: header length too large")
	ErrBodyTooLarge     = errors.New("protocol: body length too large")
	ErrInvalidHeaderBuf = errors.New("protocol: invalid fixed header buffer")
)

type FixedHeader struct {
	Magic     uint16
	Version   uint8
	MsgType   uint8
	Flags     uint16
	CodecID   uint8
	Reserved1 uint8
	Status    uint16
	Reserved2 uint16
	HeaderLen uint32
	BodyLen   uint32
	RequestID uint64
	TimeoutMs uint32
}

func (header FixedHeader) Validate() error {
	if header.Magic != MagicNumber {
		return fmt.Errorf("%w: got=%#x", ErrInvalidMagic, header.Magic)
	}
	if header.Version != CurrentVersion {
		return fmt.Errorf("%w: got=%d", ErrUnsupportedVer, header.Version)
	}
	if !isValidMsgType(header.MsgType) {
		return fmt.Errorf("%w: got=%d", ErrInvalidMsgType, header.MsgType)
	}
	if header.HeaderLen > MaxHeaderLen {
		return fmt.Errorf("%w: got=%d max=%d", ErrHeaderTooLarge, header.HeaderLen, MaxHeaderLen)
	}
	if header.BodyLen > MaxBodyLen {
		return fmt.Errorf("%w: got=%d max=%d", ErrBodyTooLarge, header.BodyLen, MaxBodyLen)
	}
	return nil
}

func (header FixedHeader) Encode() ([]byte, error) {
	if err := header.Validate(); err != nil {
		return nil, err
	}
	buf := make([]byte, FixedHeaderSize)
	binary.BigEndian.PutUint16(buf[0:2], header.Magic)
	buf[2] = header.Version
	buf[3] = header.MsgType
	binary.BigEndian.PutUint16(buf[4:6], header.Flags)
	buf[6] = header.CodecID
	buf[7] = header.Reserved1
	binary.BigEndian.PutUint16(buf[8:10], header.Status)
	binary.BigEndian.PutUint16(buf[10:12], header.Reserved2)
	binary.BigEndian.PutUint32(buf[12:16], header.HeaderLen)
	binary.BigEndian.PutUint32(buf[16:20], header.BodyLen)
	binary.BigEndian.PutUint64(buf[20:28], header.RequestID)
	binary.BigEndian.PutUint32(buf[28:32], header.TimeoutMs)
	return buf, nil
}

func DecodeFixedHeader(buf []byte) (FixedHeader, error) {
	if len(buf) != FixedHeaderSize {
		return FixedHeader{}, fmt.Errorf("%w: got=%d want=%d", ErrInvalidHeaderBuf, len(buf), FixedHeaderSize)
	}
	header := FixedHeader{
		Magic:     binary.BigEndian.Uint16(buf[0:2]),
		Version:   buf[2],
		MsgType:   buf[3],
		Flags:     binary.BigEndian.Uint16(buf[4:6]),
		CodecID:   buf[6],
		Reserved1: buf[7],
		Status:    binary.BigEndian.Uint16(buf[8:10]),
		Reserved2: binary.BigEndian.Uint16(buf[10:12]),
		HeaderLen: binary.BigEndian.Uint32(buf[12:16]),
		BodyLen:   binary.BigEndian.Uint32(buf[16:20]),
		RequestID: binary.BigEndian.Uint64(buf[20:28]),
		TimeoutMs: binary.BigEndian.Uint32(buf[28:32]),
	}
	if err := header.Validate(); err != nil {
		return FixedHeader{}, err
	}
	return header, nil
}

func isValidMsgType(msgType uint8) bool {
	switch msgType {
	case MsgTypeRequest, MsgTypeResponse, MsgTypePing, MsgTypePong:
		return true
	default:
		return false
	}
}
