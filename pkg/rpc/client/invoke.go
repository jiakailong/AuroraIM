package client

import (
	"context"
	"fmt"
	"math"
	"time"

	"kama_chat_server/pkg/rpc/api"
	"kama_chat_server/pkg/rpc/codec"
	rpcerrors "kama_chat_server/pkg/rpc/errors"
	"kama_chat_server/pkg/rpc/protocol"
)

func (client *Client) Invoke(ctx context.Context, method string, request any, response any) error {
	if client.isClosed() {
		return ErrClientClosed
	}
	if method == "" {
		return rpcerrors.New(rpcerrors.BadRequest, "method is empty", nil)
	}

	invokeCtx, cancel := client.withTimeoutContext(ctx)
	defer cancel()

	codecImpl, err := client.codecByID(client.options.CodecID)
	if err != nil {
		return err
	}

	body, err := client.marshalRequest(codecImpl, request)
	if err != nil {
		return err
	}

	attempt := 0
	for {
		attemptAddresses, pickErr := client.pickAttemptAddresses(method, 2)
		if pickErr != nil {
			return pickErr
		}

		var lastErr error
		for index, address := range attemptAddresses {
			invokeErr := client.invokeOnAddress(invokeCtx, address, method, body, codecImpl, response)
			if invokeErr == nil {
				return nil
			}
			lastErr = invokeErr
			if index == len(attemptAddresses)-1 || !shouldFailover(invokeErr) {
				break
			}
		}

		if !client.canRetry(method, request, lastErr, attempt) {
			if lastErr != nil {
				return lastErr
			}
			return rpcerrors.New(rpcerrors.Unavailable, "invoke failed without detail", nil)
		}

		delay := client.nextRetryDelay(attempt)
		if ok := sleepWithContext(invokeCtx.Done(), delay); !ok {
			return rpcerrors.New(rpcerrors.Timeout, invokeCtx.Err().Error(), nil)
		}
		attempt++
	}
}

func (client *Client) invokeOnAddress(
	invokeCtx context.Context,
	address string,
	method string,
	body []byte,
	codecImpl codec.Codec,
	response any,
) error {
	metadata := toProtocolMetadata(api.MetadataFromContext(invokeCtx))
	varHeader, err := protocol.EncodeVarHeader(method, metadata)
	if err != nil {
		return rpcerrors.New(rpcerrors.BadRequest, fmt.Sprintf("encode var header failed: %v", err), nil)
	}

	requestID := client.nextRequestID()
	requestFrame := protocol.Frame{
		Header: protocol.FixedHeader{
			Version:   protocol.CurrentVersion,
			MsgType:   protocol.MsgTypeRequest,
			CodecID:   client.options.CodecID,
			Status:    protocol.StatusOK,
			RequestID: requestID,
			TimeoutMs: timeoutMsFromContext(invokeCtx),
		},
		VarHeader: varHeader,
		Body:      body,
	}

	pool, err := client.getConnPool(address)
	if err != nil {
		return rpcerrors.New(rpcerrors.Unavailable, fmt.Sprintf("acquire pool failed: %v", err), nil)
	}

	conn, err := pool.Get()
	if err != nil {
		return rpcerrors.New(rpcerrors.Unavailable, fmt.Sprintf("acquire pooled connection failed: %v", err), nil)
	}
	shouldDiscard := false
	defer func() {
		if shouldDiscard {
			pool.Discard(conn)
			return
		}
		pool.Put(conn)
	}()

	resultCh := make(chan pendingResult, 1)
	client.addPending(requestID, resultCh)
	if err = conn.WriteFrame(requestFrame); err != nil {
		client.removePending(requestID)
		shouldDiscard = true
		return rpcerrors.New(rpcerrors.Unavailable, fmt.Sprintf("write request failed: %v", err), nil)
	}

	select {
	case <-invokeCtx.Done():
		client.removePending(requestID)
		return rpcerrors.New(rpcerrors.Timeout, invokeCtx.Err().Error(), nil)
	case result := <-resultCh:
		client.removePending(requestID)
		if result.err != nil {
			shouldDiscard = true
			return rpcerrors.New(rpcerrors.Unavailable, result.err.Error(), nil)
		}
		if err = rpcerrors.FromProtocolResponse(result.frame.Header.Status, result.frame.Body); err != nil {
			if result.frame.Header.Status == uint16(rpcerrors.Unavailable) || result.frame.Header.Status == uint16(rpcerrors.Timeout) {
				shouldDiscard = true
			}
			return err
		}
		if response != nil && len(result.frame.Body) > 0 {
			if err = codecImpl.Unmarshal(result.frame.Body, response); err != nil {
				return rpcerrors.New(rpcerrors.Internal, fmt.Sprintf("decode response body failed: %v", err), nil)
			}
		}
		return nil
	}
}

func shouldFailover(err error) bool {
	rpcErr, ok := err.(*rpcerrors.RpcError)
	if !ok || rpcErr == nil {
		return false
	}
	return rpcErr.Code == rpcerrors.Unavailable || rpcErr.Code == rpcerrors.Timeout
}

func (client *Client) withTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, func() {}
	}
	if client.options.Timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, client.options.Timeout)
}

func (client *Client) codecByID(codecID uint8) (codec.Codec, error) {
	switch codecID {
	case 0, protocol.CodecJSONID:
		return codec.NewJSONCodec(), nil
	case protocol.CodecProtobufID:
		return codec.NewProtoCodec(), nil
	default:
		return nil, rpcerrors.New(rpcerrors.BadRequest, fmt.Sprintf("unsupported codec id: %d", codecID), nil)
	}
}

func (client *Client) marshalRequest(codecImpl codec.Codec, request any) ([]byte, error) {
	if request == nil {
		return nil, nil
	}
	body, err := codecImpl.Marshal(request)
	if err != nil {
		return nil, rpcerrors.New(rpcerrors.BadRequest, fmt.Sprintf("encode request body failed: %v", err), nil)
	}
	return body, nil
}

func timeoutMsFromContext(ctx context.Context) uint32 {
	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		return 0
	}
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return 0
	}
	milliseconds := remaining.Milliseconds()
	if milliseconds > int64(math.MaxUint32) {
		return math.MaxUint32
	}
	return uint32(milliseconds)
}

func toProtocolMetadata(metadata api.Metadata) protocol.Metadata {
	if len(metadata) == 0 {
		return nil
	}
	protocolMetadata := make(protocol.Metadata, len(metadata))
	for key, value := range metadata {
		protocolMetadata[key] = value
	}
	return protocolMetadata
}
