package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"kama_chat_server/pkg/rpc/api"
)

func EnsureTraceID(ctx context.Context) (context.Context, string) {
	if traceID, ok := TraceIDFromContext(ctx); ok && traceID != "" {
		return ctx, traceID
	}
	traceID := newTraceID()
	return api.SetTraceID(ctx, traceID), traceID
}

func TraceIDFromContext(ctx context.Context) (string, bool) {
	return api.TraceIDFromContext(ctx)
}

func newTraceID() string {
	buffer := make([]byte, 8)
	if _, err := rand.Read(buffer); err == nil {
		return hex.EncodeToString(buffer)
	}
	return fmt.Sprintf("trace-%d", time.Now().UnixNano())
}
