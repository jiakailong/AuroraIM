package tests

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	rpcerrors "kama_chat_server/pkg/rpc/errors"
	"kama_chat_server/pkg/rpc/protocol"
	rpcserver "kama_chat_server/pkg/rpc/server"
	"kama_chat_server/pkg/rpc/transport"
)

type helloService struct{}

type helloRequest struct {
	Name string `json:"name"`
}

type helloResponse struct {
	Message string `json:"message"`
}

func (service *helloService) Hello(ctx context.Context, request *helloRequest) (*helloResponse, error) {
	return &helloResponse{Message: "hello," + request.Name}, nil
}

type panicService struct{}

type panicRequest struct {
	Input string `json:"input"`
}

type panicResponse struct {
	Result string `json:"result"`
}

func (service *panicService) Panic(ctx context.Context, request *panicRequest) (*panicResponse, error) {
	panic("boom")
}

func TestClientServer(t *testing.T) {
	srv := rpcserver.NewServer()
	srv.Register("Echo.Say", func(ctx context.Context, request *protocol.Request) (*protocol.Response, error) {
		return &protocol.Response{
			Status: protocol.StatusOK,
			Body:   append([]byte("echo:"), request.Body...),
		}, nil
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer func() {
		_ = srv.Close()
	}()
	go func() {
		_ = srv.Serve(listener)
	}()

	rawConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	resultCh := make(chan protocol.Frame, 1)
	clientConn := transport.NewConn(rawConn, func(conn *transport.Conn, frame protocol.Frame) {
		resultCh <- frame
	}, func(conn *transport.Conn, err error) {})
	clientConn.Start()
	defer func() {
		_ = clientConn.Close()
	}()

	varHeader, err := protocol.EncodeVarHeader("Echo.Say", protocol.Metadata{"trace_id": "day5-test"})
	if err != nil {
		t.Fatalf("encode var header failed: %v", err)
	}

	requestFrame := protocol.Frame{
		Header: protocol.FixedHeader{
			Version:   protocol.CurrentVersion,
			MsgType:   protocol.MsgTypeRequest,
			CodecID:   protocol.DefaultCodecID,
			Status:    protocol.StatusOK,
			RequestID: 1001,
			TimeoutMs: 1000,
		},
		VarHeader: varHeader,
		Body:      []byte("hello"),
	}
	if err = clientConn.WriteFrame(requestFrame); err != nil {
		t.Fatalf("write request frame failed: %v", err)
	}

	select {
	case responseFrame := <-resultCh:
		if responseFrame.Header.MsgType != protocol.MsgTypeResponse {
			t.Fatalf("unexpected msg type: got=%d", responseFrame.Header.MsgType)
		}
		if responseFrame.Header.Status != protocol.StatusOK {
			t.Fatalf("unexpected status: got=%d", responseFrame.Header.Status)
		}
		if string(responseFrame.Body) != "echo:hello" {
			t.Fatalf("unexpected response body: got=%s", string(responseFrame.Body))
		}
	case <-time.After(3 * time.Second):
		t.Fatal("wait response timeout")
	}
}

func TestClientServerMethodNotFound(t *testing.T) {
	srv := rpcserver.NewServer()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer func() {
		_ = srv.Close()
	}()
	go func() {
		_ = srv.Serve(listener)
	}()

	rawConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	resultCh := make(chan protocol.Frame, 1)
	clientConn := transport.NewConn(rawConn, func(conn *transport.Conn, frame protocol.Frame) {
		resultCh <- frame
	}, func(conn *transport.Conn, err error) {})
	clientConn.Start()
	defer func() {
		_ = clientConn.Close()
	}()

	varHeader, err := protocol.EncodeVarHeader("Unknown.Method", nil)
	if err != nil {
		t.Fatalf("encode var header failed: %v", err)
	}
	requestFrame := protocol.Frame{
		Header: protocol.FixedHeader{
			Version:   protocol.CurrentVersion,
			MsgType:   protocol.MsgTypeRequest,
			CodecID:   protocol.DefaultCodecID,
			RequestID: 1002,
		},
		VarHeader: varHeader,
		Body:      []byte("x"),
	}
	if err = clientConn.WriteFrame(requestFrame); err != nil {
		t.Fatalf("write request frame failed: %v", err)
	}

	select {
	case responseFrame := <-resultCh:
		if responseFrame.Header.Status != uint16(rpcerrors.NotFound) {
			t.Fatalf("unexpected status for not found: got=%d want=%d", responseFrame.Header.Status, rpcerrors.NotFound)
		}
		if convertErr := rpcerrors.FromProtocolResponse(responseFrame.Header.Status, responseFrame.Body); convertErr == nil {
			t.Fatal("expected converted rpc error for method not found")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("wait response timeout")
	}
}

func TestClientServerRegisterServiceReflect(t *testing.T) {
	srv := rpcserver.NewServer()
	if err := srv.RegisterService(&helloService{}); err != nil {
		t.Fatalf("register service by reflection failed: %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer func() {
		_ = srv.Close()
	}()
	go func() {
		_ = srv.Serve(listener)
	}()

	rawConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	resultCh := make(chan protocol.Frame, 1)
	clientConn := transport.NewConn(rawConn, func(conn *transport.Conn, frame protocol.Frame) {
		resultCh <- frame
	}, func(conn *transport.Conn, err error) {})
	clientConn.Start()
	defer func() {
		_ = clientConn.Close()
	}()

	body, err := json.Marshal(&helloRequest{Name: "kama"})
	if err != nil {
		t.Fatalf("marshal request body failed: %v", err)
	}
	varHeader, err := protocol.EncodeVarHeader("helloService.Hello", nil)
	if err != nil {
		t.Fatalf("encode var header failed: %v", err)
	}
	requestFrame := protocol.Frame{
		Header: protocol.FixedHeader{
			Version:   protocol.CurrentVersion,
			MsgType:   protocol.MsgTypeRequest,
			CodecID:   protocol.DefaultCodecID,
			RequestID: 2001,
		},
		VarHeader: varHeader,
		Body:      body,
	}
	if err = clientConn.WriteFrame(requestFrame); err != nil {
		t.Fatalf("write request frame failed: %v", err)
	}

	select {
	case responseFrame := <-resultCh:
		if responseFrame.Header.Status != protocol.StatusOK {
			t.Fatalf("unexpected status: got=%d", responseFrame.Header.Status)
		}
		var response helloResponse
		if err = json.Unmarshal(responseFrame.Body, &response); err != nil {
			t.Fatalf("unmarshal response body failed: %v", err)
		}
		if response.Message != "hello,kama" {
			t.Fatalf("unexpected response: %+v", response)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("wait response timeout")
	}
}

func TestClientServerHandlerPanicRecovered(t *testing.T) {
	srv := rpcserver.NewServer()
	if err := srv.RegisterService(&panicService{}); err != nil {
		t.Fatalf("register panic service failed: %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer func() {
		_ = srv.Close()
	}()
	go func() {
		_ = srv.Serve(listener)
	}()

	rawConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	resultCh := make(chan protocol.Frame, 1)
	clientConn := transport.NewConn(rawConn, func(conn *transport.Conn, frame protocol.Frame) {
		resultCh <- frame
	}, func(conn *transport.Conn, err error) {})
	clientConn.Start()
	defer func() {
		_ = clientConn.Close()
	}()

	body, err := json.Marshal(&panicRequest{Input: "x"})
	if err != nil {
		t.Fatalf("marshal request body failed: %v", err)
	}
	varHeader, err := protocol.EncodeVarHeader("panicService.Panic", nil)
	if err != nil {
		t.Fatalf("encode var header failed: %v", err)
	}
	requestFrame := protocol.Frame{
		Header: protocol.FixedHeader{
			Version:   protocol.CurrentVersion,
			MsgType:   protocol.MsgTypeRequest,
			CodecID:   protocol.DefaultCodecID,
			RequestID: 2002,
		},
		VarHeader: varHeader,
		Body:      body,
	}
	if err = clientConn.WriteFrame(requestFrame); err != nil {
		t.Fatalf("write request frame failed: %v", err)
	}

	select {
	case responseFrame := <-resultCh:
		if responseFrame.Header.Status != uint16(rpcerrors.Internal) {
			t.Fatalf("expected internal status, got=%d", responseFrame.Header.Status)
		}
		if convertErr := rpcerrors.FromProtocolResponse(responseFrame.Header.Status, responseFrame.Body); convertErr == nil {
			t.Fatal("expected converted error for panic")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("wait response timeout")
	}
}
