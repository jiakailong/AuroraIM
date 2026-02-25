package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type Metadata map[string]string

type Request struct {
	RequestID uint64
	Method    string
	Metadata  Metadata
	CodecID   uint8
	Flags     uint16
	TimeoutMs uint32
	Body      []byte
}

type Response struct {
	RequestID uint64
	Metadata  Metadata
	CodecID   uint8
	Flags     uint16
	Status    uint16
	Body      []byte
}

func EncodeVarHeader(method string, metadata Metadata) ([]byte, error) {
	if len(method) > int(^uint16(0)) {
		return nil, fmt.Errorf("protocol: method too long: %d", len(method))
	}
	if len(metadata) > int(^uint16(0)) {
		return nil, fmt.Errorf("protocol: too many metadata entries: %d", len(metadata))
	}

	var buffer bytes.Buffer
	if err := binary.Write(&buffer, binary.BigEndian, uint16(len(method))); err != nil {
		return nil, fmt.Errorf("protocol: write method length: %w", err)
	}
	if _, err := buffer.WriteString(method); err != nil {
		return nil, fmt.Errorf("protocol: write method: %w", err)
	}
	if err := binary.Write(&buffer, binary.BigEndian, uint16(len(metadata))); err != nil {
		return nil, fmt.Errorf("protocol: write metadata count: %w", err)
	}
	for key, value := range metadata {
		if len(key) > int(^uint16(0)) || len(value) > int(^uint16(0)) {
			return nil, fmt.Errorf("protocol: metadata key/value too long")
		}
		if err := binary.Write(&buffer, binary.BigEndian, uint16(len(key))); err != nil {
			return nil, fmt.Errorf("protocol: write metadata key len: %w", err)
		}
		if _, err := buffer.WriteString(key); err != nil {
			return nil, fmt.Errorf("protocol: write metadata key: %w", err)
		}
		if err := binary.Write(&buffer, binary.BigEndian, uint16(len(value))); err != nil {
			return nil, fmt.Errorf("protocol: write metadata value len: %w", err)
		}
		if _, err := buffer.WriteString(value); err != nil {
			return nil, fmt.Errorf("protocol: write metadata value: %w", err)
		}
	}
	return buffer.Bytes(), nil
}

func DecodeVarHeader(data []byte) (string, Metadata, error) {
	reader := bytes.NewReader(data)

	var methodLen uint16
	if err := binary.Read(reader, binary.BigEndian, &methodLen); err != nil {
		return "", nil, fmt.Errorf("protocol: read method len: %w", err)
	}
	methodBytes := make([]byte, methodLen)
	if _, err := io.ReadFull(reader, methodBytes); err != nil {
		return "", nil, fmt.Errorf("protocol: read method: %w", err)
	}

	var metaCount uint16
	if err := binary.Read(reader, binary.BigEndian, &metaCount); err != nil {
		return "", nil, fmt.Errorf("protocol: read metadata count: %w", err)
	}
	metadata := make(Metadata, metaCount)
	for index := 0; index < int(metaCount); index++ {
		var keyLen uint16
		if err := binary.Read(reader, binary.BigEndian, &keyLen); err != nil {
			return "", nil, fmt.Errorf("protocol: read metadata key len: %w", err)
		}
		key := make([]byte, keyLen)
		if _, err := io.ReadFull(reader, key); err != nil {
			return "", nil, fmt.Errorf("protocol: read metadata key: %w", err)
		}

		var valueLen uint16
		if err := binary.Read(reader, binary.BigEndian, &valueLen); err != nil {
			return "", nil, fmt.Errorf("protocol: read metadata value len: %w", err)
		}
		value := make([]byte, valueLen)
		if _, err := io.ReadFull(reader, value); err != nil {
			return "", nil, fmt.Errorf("protocol: read metadata value: %w", err)
		}
		metadata[string(key)] = string(value)
	}
	if reader.Len() != 0 {
		return "", nil, fmt.Errorf("protocol: trailing bytes in var header: %d", reader.Len())
	}

	return string(methodBytes), metadata, nil
}
