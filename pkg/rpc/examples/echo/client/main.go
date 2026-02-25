package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"kama_chat_server/pkg/rpc/protocol"
	"kama_chat_server/pkg/rpc/transport"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:19090", "echo server address")
	flag.Parse()

	rawConn, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatalf("dial echo server failed: %v", err)
	}

	done := make(chan struct{})
	conn := transport.NewConn(rawConn, func(conn *transport.Conn, frame protocol.Frame) {
		method, metadata, decodeErr := protocol.DecodeVarHeader(frame.VarHeader)
		if decodeErr != nil {
			log.Printf("decode var header failed: %v", decodeErr)
		} else {
			fmt.Printf("echo response request_id=%d method=%s metadata=%v body=%s\n", frame.Header.RequestID, method, metadata, string(frame.Body))
		}
		close(done)
	}, func(conn *transport.Conn, closeErr error) {
		if closeErr != nil {
			log.Printf("client connection closed with error: %v", closeErr)
		}
	})
	conn.Start()
	defer func() {
		_ = conn.Close()
	}()

	varHeader, err := protocol.EncodeVarHeader("Echo.Ping", protocol.Metadata{"trace_id": "echo-trace-1"})
	if err != nil {
		log.Fatalf("encode var header failed: %v", err)
	}

	requestFrame := protocol.Frame{
		Header: protocol.FixedHeader{
			Version:   protocol.CurrentVersion,
			MsgType:   protocol.MsgTypeRequest,
			CodecID:   protocol.DefaultCodecID,
			Status:    protocol.StatusOK,
			RequestID: 1,
			TimeoutMs: 1000,
		},
		VarHeader: varHeader,
		Body:      []byte("hello echo"),
	}
	if err = conn.WriteFrame(requestFrame); err != nil {
		log.Fatalf("write request frame failed: %v", err)
	}

	select {
	case <-done:
		return
	case <-time.After(5 * time.Second):
		log.Fatal("wait echo response timeout")
	}
}
