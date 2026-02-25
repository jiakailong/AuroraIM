package server

import (
	"context"
	"fmt"
	"reflect"

	"kama_chat_server/pkg/rpc/codec"
	rpcerrors "kama_chat_server/pkg/rpc/errors"
	"kama_chat_server/pkg/rpc/protocol"
)

// HandlerFunc 定义 RPC 方法处理函数。
type HandlerFunc func(ctx context.Context, request *protocol.Request) (*protocol.Response, error)

func buildReflectedHandler(receiver reflect.Value, method reflect.Method, requestType reflect.Type, responseType reflect.Type) (HandlerFunc, error) {
	return func(ctx context.Context, request *protocol.Request) (*protocol.Response, error) {
		codecImpl, err := codecByID(request.CodecID)
		if err != nil {
			return nil, err
		}

		requestValue := reflect.New(requestType.Elem())
		if len(request.Body) > 0 {
			if err = codecImpl.Unmarshal(request.Body, requestValue.Interface()); err != nil {
				return nil, rpcerrors.New(rpcerrors.BadRequest, fmt.Sprintf("decode request body failed: %v", err), nil)
			}
		}

		results := method.Func.Call([]reflect.Value{receiver, reflect.ValueOf(ctx), requestValue})
		errorValue := results[1]
		if !errorValue.IsNil() {
			return nil, errorValue.Interface().(error)
		}

		responseValue := results[0]
		if responseValue.Type() != responseType {
			return nil, rpcerrors.New(rpcerrors.Internal, "unexpected response type from reflected handler", nil)
		}

		responseBody, err := codecImpl.Marshal(responseValue.Interface())
		if err != nil {
			return nil, rpcerrors.New(rpcerrors.Internal, fmt.Sprintf("encode response body failed: %v", err), nil)
		}

		return &protocol.Response{
			RequestID: request.RequestID,
			CodecID:   request.CodecID,
			Status:    protocol.StatusOK,
			Body:      responseBody,
		}, nil
	}, nil
}

func codecByID(codecID uint8) (codec.Codec, error) {
	switch codecID {
	case 0, protocol.CodecJSONID:
		return codec.NewJSONCodec(), nil
	case protocol.CodecProtobufID:
		return codec.NewProtoCodec(), nil
	default:
		return nil, rpcerrors.New(rpcerrors.BadRequest, fmt.Sprintf("unsupported codec id: %d", codecID), nil)
	}
}
