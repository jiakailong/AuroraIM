package api

import "context"

const (
	MetadataKeyTraceID   = "trace_id"
	MetadataKeyDeadline  = "deadline"
	MetadataKeyAuthToken = "auth_token"
)

type Metadata map[string]string

type metadataContextKey struct{}

func WithMetadata(ctx context.Context, metadata Metadata) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	copyMetadata := cloneMetadata(metadata)
	return context.WithValue(ctx, metadataContextKey{}, copyMetadata)
}

func MetadataFromContext(ctx context.Context) Metadata {
	if ctx == nil {
		return Metadata{}
	}
	metadata, ok := ctx.Value(metadataContextKey{}).(Metadata)
	if !ok {
		return Metadata{}
	}
	return cloneMetadata(metadata)
}

func SetMetadataValue(ctx context.Context, key, value string) context.Context {
	metadata := MetadataFromContext(ctx)
	metadata[key] = value
	return WithMetadata(ctx, metadata)
}

func GetMetadataValue(ctx context.Context, key string) (string, bool) {
	metadata := MetadataFromContext(ctx)
	value, ok := metadata[key]
	return value, ok
}

func SetTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		return ctx
	}
	return SetMetadataValue(ctx, MetadataKeyTraceID, traceID)
}

func TraceIDFromContext(ctx context.Context) (string, bool) {
	return GetMetadataValue(ctx, MetadataKeyTraceID)
}

func cloneMetadata(metadata Metadata) Metadata {
	if len(metadata) == 0 {
		return Metadata{}
	}
	copied := make(Metadata, len(metadata))
	for key, value := range metadata {
		copied[key] = value
	}
	return copied
}
