package tests

import (
	"encoding/json"
	"testing"

	rpcerrors "kama_chat_server/pkg/rpc/errors"
)

func TestError(t *testing.T) {
	t.Run("rpc_error_json_roundtrip", func(t *testing.T) {
		want := rpcerrors.New(rpcerrors.BadRequest, "invalid payload", map[string]any{"field": "method"})
		encoded, err := json.Marshal(want)
		if err != nil {
			t.Fatalf("marshal rpc error failed: %v", err)
		}

		var got rpcerrors.RpcError
		if err = json.Unmarshal(encoded, &got); err != nil {
			t.Fatalf("unmarshal rpc error failed: %v", err)
		}
		if got.Code != want.Code {
			t.Fatalf("error code mismatch: got=%d want=%d", got.Code, want.Code)
		}
		if got.Message != want.Message {
			t.Fatalf("error message mismatch: got=%s want=%s", got.Message, want.Message)
		}
	})

	t.Run("from_protocol_response_with_json_body", func(t *testing.T) {
		body, err := json.Marshal(rpcerrors.New(rpcerrors.NotFound, "method not found", nil))
		if err != nil {
			t.Fatalf("marshal rpc error body failed: %v", err)
		}

		convertedErr := rpcerrors.FromProtocolResponse(uint16(rpcerrors.NotFound), body)
		if convertedErr == nil {
			t.Fatal("expected converted error, got nil")
		}
		rpcErr, ok := convertedErr.(*rpcerrors.RpcError)
		if !ok {
			t.Fatalf("expected *RpcError, got %T", convertedErr)
		}
		if rpcErr.Code != rpcerrors.NotFound {
			t.Fatalf("converted error code mismatch: got=%d want=%d", rpcErr.Code, rpcerrors.NotFound)
		}
		if rpcErr.Message != "method not found" {
			t.Fatalf("converted error message mismatch: got=%s", rpcErr.Message)
		}
	})

	t.Run("from_protocol_response_with_plain_body", func(t *testing.T) {
		convertedErr := rpcerrors.FromProtocolResponse(uint16(rpcerrors.Timeout), []byte("request timed out"))
		if convertedErr == nil {
			t.Fatal("expected timeout error, got nil")
		}
		rpcErr, ok := convertedErr.(*rpcerrors.RpcError)
		if !ok {
			t.Fatalf("expected *RpcError, got %T", convertedErr)
		}
		if rpcErr.Code != rpcerrors.Timeout {
			t.Fatalf("timeout code mismatch: got=%d want=%d", rpcErr.Code, rpcerrors.Timeout)
		}
		if rpcErr.Message != "request timed out" {
			t.Fatalf("timeout message mismatch: got=%s", rpcErr.Message)
		}
	})

	t.Run("from_protocol_response_ok", func(t *testing.T) {
		if convertedErr := rpcerrors.FromProtocolResponse(uint16(rpcerrors.OK), nil); convertedErr != nil {
			t.Fatalf("expected nil for success status, got %v", convertedErr)
		}
	})
}
