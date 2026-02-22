package protocol

import (
	"fmt"
	"io"
)

type Frame struct {
	Header    FixedHeader
	VarHeader []byte
	Body      []byte
}

func WriteFrame(writer io.Writer, frame Frame) error {
	frame.Header.Magic = MagicNumber
	if frame.Header.Version == 0 {
		frame.Header.Version = CurrentVersion
	}
	frame.Header.HeaderLen = uint32(len(frame.VarHeader))
	frame.Header.BodyLen = uint32(len(frame.Body))

	headerBytes, err := frame.Header.Encode()
	if err != nil {
		return err
	}
	if _, err = writer.Write(headerBytes); err != nil {
		return fmt.Errorf("protocol: write fixed header: %w", err)
	}
	if len(frame.VarHeader) > 0 {
		if _, err = writer.Write(frame.VarHeader); err != nil {
			return fmt.Errorf("protocol: write var header: %w", err)
		}
	}
	if len(frame.Body) > 0 {
		if _, err = writer.Write(frame.Body); err != nil {
			return fmt.Errorf("protocol: write body: %w", err)
		}
	}
	return nil
}

func ReadFrame(reader io.Reader) (Frame, error) {
	fixedHeaderBytes := make([]byte, FixedHeaderSize)
	if _, err := io.ReadFull(reader, fixedHeaderBytes); err != nil {
		return Frame{}, fmt.Errorf("protocol: read fixed header: %w", err)
	}
	header, err := DecodeFixedHeader(fixedHeaderBytes)
	if err != nil {
		return Frame{}, err
	}

	frame := Frame{Header: header}
	if header.HeaderLen > 0 {
		frame.VarHeader = make([]byte, header.HeaderLen)
		if _, err = io.ReadFull(reader, frame.VarHeader); err != nil {
			return Frame{}, fmt.Errorf("protocol: read var header: %w", err)
		}
	}
	if header.BodyLen > 0 {
		frame.Body = make([]byte, header.BodyLen)
		if _, err = io.ReadFull(reader, frame.Body); err != nil {
			return Frame{}, fmt.Errorf("protocol: read body: %w", err)
		}
	}
	return frame, nil
}
