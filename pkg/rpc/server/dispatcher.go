package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"kama_chat_server/pkg/rpc/api"
	rpcerrors "kama_chat_server/pkg/rpc/errors"
	"kama_chat_server/pkg/rpc/middleware"
	"kama_chat_server/pkg/rpc/observability"
	"kama_chat_server/pkg/rpc/protocol"
	"kama_chat_server/pkg/rpc/transport"
)

type Dispatcher struct {
	mu       sync.RWMutex
	handlers map[string]HandlerFunc
	chain    middleware.UnaryMiddleware
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string]HandlerFunc),
		chain: middleware.Chain(
			middleware.Tracing(),
			middleware.Metrics(observability.GetDefaultMetrics()),
			middleware.Logging(observability.GetDefaultLogger()),
			middleware.Recovery(),
		),
	}
}

func (dispatcher *Dispatcher) Register(method string, handler HandlerFunc) {
	dispatcher.mu.Lock()
	defer dispatcher.mu.Unlock()
	dispatcher.handlers[method] = handler
}

func (dispatcher *Dispatcher) Dispatch(conn *transport.Conn, frame protocol.Frame) {
	if frame.Header.MsgType != protocol.MsgTypeRequest {
		return
	}

	method, metadata, err := protocol.DecodeVarHeader(frame.VarHeader)
	if err != nil {
		dispatcher.writeError(conn, frame, "", nil, rpcerrors.New(rpcerrors.BadRequest, fmt.Sprintf("decode var header failed: %v", err), nil))
		return
	}

	handler, ok := dispatcher.getHandler(method)
	if !ok {
		dispatcher.writeError(conn, frame, method, metadata, rpcerrors.New(rpcerrors.NotFound, "method not found", map[string]any{"method": method}))
		return
	}

	request := &protocol.Request{
		RequestID: frame.Header.RequestID,
		Method:    method,
		Metadata:  metadata,
		CodecID:   frame.Header.CodecID,
		Flags:     frame.Header.Flags,
		TimeoutMs: frame.Header.TimeoutMs,
		Body:      frame.Body,
	}

	ctx := api.WithMetadata(context.Background(), api.Metadata(metadata))
	wrapped := dispatcher.chain(middleware.UnaryHandler(handler))
	response, handlerErr := wrapped(ctx, request)
	if handlerErr != nil {
		dispatcher.writeHandlerError(conn, frame, method, request.Metadata, handlerErr)
		return
	}
	if response == nil {
		response = &protocol.Response{}
	}
	if response.RequestID == 0 {
		response.RequestID = request.RequestID
	}
	if response.CodecID == 0 {
		response.CodecID = request.CodecID
	}
	if response.Status == 0 {
		response.Status = protocol.StatusOK
	}

	responseMetadata := response.Metadata
	if responseMetadata == nil {
		responseMetadata = request.Metadata
	}
	varHeader, err := protocol.EncodeVarHeader(method, responseMetadata)
	if err != nil {
		dispatcher.writeError(conn, frame, method, request.Metadata, rpcerrors.New(rpcerrors.Internal, fmt.Sprintf("encode response var header failed: %v", err), nil))
		return
	}

	responseFrame := protocol.Frame{
		Header: protocol.FixedHeader{
			Version:   protocol.CurrentVersion,
			MsgType:   protocol.MsgTypeResponse,
			Flags:     response.Flags,
			CodecID:   response.CodecID,
			Status:    response.Status,
			RequestID: response.RequestID,
			TimeoutMs: request.TimeoutMs,
		},
		VarHeader: varHeader,
		Body:      response.Body,
	}
	if err = conn.WriteFrame(responseFrame); err != nil {
		_ = conn.Close()
	}
}

func (dispatcher *Dispatcher) getHandler(method string) (HandlerFunc, bool) {
	dispatcher.mu.RLock()
	defer dispatcher.mu.RUnlock()
	handler, ok := dispatcher.handlers[method]
	return handler, ok
}

func (dispatcher *Dispatcher) writeHandlerError(conn *transport.Conn, requestFrame protocol.Frame, method string, metadata protocol.Metadata, handlerErr error) {
	var rpcErr *rpcerrors.RpcError
	if errors.As(handlerErr, &rpcErr) {
		dispatcher.writeError(conn, requestFrame, method, metadata, rpcErr)
		return
	}
	dispatcher.writeError(conn, requestFrame, method, metadata, rpcerrors.New(rpcerrors.Internal, handlerErr.Error(), nil))
}

func (dispatcher *Dispatcher) writeError(conn *transport.Conn, requestFrame protocol.Frame, method string, metadata protocol.Metadata, rpcErr *rpcerrors.RpcError) {
	if rpcErr == nil {
		rpcErr = rpcerrors.New(rpcerrors.Internal, "rpc internal error", nil)
	}
	body, err := json.Marshal(rpcErr)
	if err != nil {
		body = []byte(rpcErr.Error())
	}
	varHeader, err := protocol.EncodeVarHeader(method, metadata)
	if err != nil {
		varHeader, _ = protocol.EncodeVarHeader("", nil)
	}

	responseFrame := protocol.Frame{
		Header: protocol.FixedHeader{
			Version:   protocol.CurrentVersion,
			MsgType:   protocol.MsgTypeResponse,
			CodecID:   requestFrame.Header.CodecID,
			Status:    rpcErr.Code.Uint16(),
			RequestID: requestFrame.Header.RequestID,
			TimeoutMs: requestFrame.Header.TimeoutMs,
		},
		VarHeader: varHeader,
		Body:      body,
	}
	if writeErr := conn.WriteFrame(responseFrame); writeErr != nil {
		_ = conn.Close()
	}
}
