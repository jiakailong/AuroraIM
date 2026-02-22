package tests

import (
	"bytes"
	"io"
	"testing"

	"kama_chat_server/pkg/rpc/protocol"
)

func TestProtocol(t *testing.T) {
	t.Run("encode_decode_consistency", func(t *testing.T) {
		metadata := protocol.Metadata{
			"trace_id":   "trace-123",
			"auth_token": "token-abc",
		}
		varHeader, err := protocol.EncodeVarHeader("MessageService.GetMessageList", metadata)
		if err != nil {
			t.Fatalf("encode var header failed: %v", err)
		}

		wantBody := []byte(`{"user_one_id":"U1","user_two_id":"U2"}`)
		wantFrame := protocol.Frame{
			Header: protocol.FixedHeader{
				Version:   protocol.CurrentVersion,
				MsgType:   protocol.MsgTypeRequest,
				Flags:     0,
				CodecID:   protocol.DefaultCodecID,
				Status:    protocol.StatusOK,
				RequestID: 101,
				TimeoutMs: 1200,
			},
			VarHeader: varHeader,
			Body:      wantBody,
		}

		buffer := bytes.NewBuffer(nil)
		if err = protocol.WriteFrame(buffer, wantFrame); err != nil {
			t.Fatalf("write frame failed: %v", err)
		}

		gotFrame, err := protocol.ReadFrame(buffer)
		if err != nil {
			t.Fatalf("read frame failed: %v", err)
		}

		if gotFrame.Header.RequestID != wantFrame.Header.RequestID {
			t.Fatalf("request id mismatch: got=%d want=%d", gotFrame.Header.RequestID, wantFrame.Header.RequestID)
		}
		if gotFrame.Header.TimeoutMs != wantFrame.Header.TimeoutMs {
			t.Fatalf("timeout mismatch: got=%d want=%d", gotFrame.Header.TimeoutMs, wantFrame.Header.TimeoutMs)
		}
		if gotFrame.Header.MsgType != protocol.MsgTypeRequest {
			t.Fatalf("msg type mismatch: got=%d", gotFrame.Header.MsgType)
		}
		if !bytes.Equal(gotFrame.Body, wantBody) {
			t.Fatalf("body mismatch: got=%s want=%s", string(gotFrame.Body), string(wantBody))
		}

		method, gotMetadata, err := protocol.DecodeVarHeader(gotFrame.VarHeader)
		if err != nil {
			t.Fatalf("decode var header failed: %v", err)
		}
		if method != "MessageService.GetMessageList" {
			t.Fatalf("method mismatch: got=%s", method)
		}
		if gotMetadata["trace_id"] != metadata["trace_id"] || gotMetadata["auth_token"] != metadata["auth_token"] {
			t.Fatalf("metadata mismatch: got=%v want=%v", gotMetadata, metadata)
		}
	})

	t.Run("sticky_packets", func(t *testing.T) {
		frame1 := buildResponseFrame(201, []byte(`{"ok":true,"seq":1}`))
		frame2 := buildResponseFrame(202, []byte(`{"ok":true,"seq":2}`))

		stream := bytes.NewBuffer(nil)
		if err := protocol.WriteFrame(stream, frame1); err != nil {
			t.Fatalf("write frame1 failed: %v", err)
		}
		if err := protocol.WriteFrame(stream, frame2); err != nil {
			t.Fatalf("write frame2 failed: %v", err)
		}

		got1, err := protocol.ReadFrame(stream)
		if err != nil {
			t.Fatalf("read frame1 failed: %v", err)
		}
		got2, err := protocol.ReadFrame(stream)
		if err != nil {
			t.Fatalf("read frame2 failed: %v", err)
		}

		if got1.Header.RequestID != 201 || got2.Header.RequestID != 202 {
			t.Fatalf("sticky packet split failed: got1=%d got2=%d", got1.Header.RequestID, got2.Header.RequestID)
		}
	})

	t.Run("half_packets", func(t *testing.T) {
		frame := buildResponseFrame(301, []byte(`{"msg":"split-read"}`))

		stream := bytes.NewBuffer(nil)
		if err := protocol.WriteFrame(stream, frame); err != nil {
			t.Fatalf("write frame failed: %v", err)
		}

		segReader := &segmentReader{
			data:   stream.Bytes(),
			steps:  []int{1, 2, 3, 1, 4, 2, 5, 1, 8, 13},
			cursor: 0,
		}
		gotFrame, err := protocol.ReadFrame(segReader)
		if err != nil {
			t.Fatalf("read half packet failed: %v", err)
		}

		if gotFrame.Header.RequestID != frame.Header.RequestID {
			t.Fatalf("request id mismatch: got=%d want=%d", gotFrame.Header.RequestID, frame.Header.RequestID)
		}
		if !bytes.Equal(gotFrame.Body, frame.Body) {
			t.Fatalf("body mismatch: got=%s want=%s", string(gotFrame.Body), string(frame.Body))
		}
	})
}

func buildResponseFrame(requestID uint64, body []byte) protocol.Frame {
	varHeader, _ := protocol.EncodeVarHeader("", nil)
	return protocol.Frame{
		Header: protocol.FixedHeader{
			Version:   protocol.CurrentVersion,
			MsgType:   protocol.MsgTypeResponse,
			CodecID:   protocol.DefaultCodecID,
			Status:    protocol.StatusOK,
			RequestID: requestID,
		},
		VarHeader: varHeader,
		Body:      body,
	}
}

type segmentReader struct {
	data   []byte
	steps  []int
	cursor int
	index  int
}

func (reader *segmentReader) Read(p []byte) (int, error) {
	if reader.cursor >= len(reader.data) {
		return 0, io.EOF
	}
	step := reader.steps[reader.index%len(reader.steps)]
	reader.index++
	if step > len(p) {
		step = len(p)
	}
	remaining := len(reader.data) - reader.cursor
	if step > remaining {
		step = remaining
	}
	copy(p[:step], reader.data[reader.cursor:reader.cursor+step])
	reader.cursor += step
	return step, nil
}
