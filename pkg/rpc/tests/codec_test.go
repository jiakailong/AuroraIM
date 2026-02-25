package tests

import (
	"errors"
	"testing"

	"kama_chat_server/pkg/rpc/codec"
	"kama_chat_server/pkg/rpc/protocol"
)

type messageListRequestPayload struct {
	UserOneID string `json:"user_one_id"`
	UserTwoID string `json:"user_two_id"`
	Page      int    `json:"page"`
}

type messageItem struct {
	SendID    string `json:"send_id"`
	ReceiveID string `json:"receive_id"`
	Content   string `json:"content"`
}

type messageListResponsePayload struct {
	Total int           `json:"total"`
	Items []messageItem `json:"items"`
}

func TestCodec(t *testing.T) {
	t.Run("json_request_body_roundtrip", func(t *testing.T) {
		jsonCodec := codec.NewJSONCodec()
		if jsonCodec.Name() != "json" {
			t.Fatalf("codec name mismatch: got=%s", jsonCodec.Name())
		}
		if jsonCodec.ID() != protocol.CodecJSONID {
			t.Fatalf("codec id mismatch: got=%d", jsonCodec.ID())
		}

		requestBody := messageListRequestPayload{
			UserOneID: "U1001",
			UserTwoID: "U1002",
			Page:      2,
		}
		encoded, err := jsonCodec.Marshal(requestBody)
		if err != nil {
			t.Fatalf("marshal request body failed: %v", err)
		}

		var decoded messageListRequestPayload
		if err = jsonCodec.Unmarshal(encoded, &decoded); err != nil {
			t.Fatalf("unmarshal request body failed: %v", err)
		}
		if decoded != requestBody {
			t.Fatalf("request body mismatch: got=%+v want=%+v", decoded, requestBody)
		}
	})

	t.Run("json_response_body_roundtrip", func(t *testing.T) {
		jsonCodec := codec.NewJSONCodec()
		responseBody := messageListResponsePayload{
			Total: 2,
			Items: []messageItem{
				{SendID: "U1001", ReceiveID: "U1002", Content: "hello"},
				{SendID: "U1002", ReceiveID: "U1001", Content: "hi"},
			},
		}
		encoded, err := jsonCodec.Marshal(responseBody)
		if err != nil {
			t.Fatalf("marshal response body failed: %v", err)
		}

		var decoded messageListResponsePayload
		if err = jsonCodec.Unmarshal(encoded, &decoded); err != nil {
			t.Fatalf("unmarshal response body failed: %v", err)
		}
		if decoded.Total != responseBody.Total || len(decoded.Items) != len(responseBody.Items) {
			t.Fatalf("response body summary mismatch: got=%+v want=%+v", decoded, responseBody)
		}
		for index := range responseBody.Items {
			if decoded.Items[index] != responseBody.Items[index] {
				t.Fatalf("response item mismatch at index=%d: got=%+v want=%+v", index, decoded.Items[index], responseBody.Items[index])
			}
		}
	})

	t.Run("proto_codec_placeholder", func(t *testing.T) {
		protoCodec := codec.NewProtoCodec()
		if protoCodec.Name() != "protobuf" {
			t.Fatalf("codec name mismatch: got=%s", protoCodec.Name())
		}
		if protoCodec.ID() != protocol.CodecProtobufID {
			t.Fatalf("codec id mismatch: got=%d", protoCodec.ID())
		}
		if _, err := protoCodec.Marshal(map[string]any{"k": "v"}); err == nil {
			t.Fatal("expected marshal error for proto placeholder, got nil")
		} else if !errors.Is(err, codec.ErrUnsupportedCodec) {
			t.Fatalf("expected ErrUnsupportedCodec, got=%v", err)
		}
		if err := protoCodec.Unmarshal([]byte("{}"), &messageListRequestPayload{}); err == nil {
			t.Fatal("expected unmarshal error for proto placeholder, got nil")
		} else if !errors.Is(err, codec.ErrUnsupportedCodec) {
			t.Fatalf("expected ErrUnsupportedCodec, got=%v", err)
		}
	})
}
