package observability

import (
	"fmt"
	"log"
	"sync"
)

// Fields 定义结构化日志字段。
type Fields map[string]any

// Logger 定义 RPC 可观测日志接口。
type Logger interface {
	Info(message string, fields Fields)
	Error(message string, fields Fields)
}

type stdLogger struct{}

func (stdLogger) Info(message string, fields Fields) {
	log.Printf("[INFO] %s %s", message, formatFields(fields))
}

func (stdLogger) Error(message string, fields Fields) {
	log.Printf("[ERROR] %s %s", message, formatFields(fields))
}

var (
	defaultLoggerMu sync.RWMutex
	defaultLogger   Logger = stdLogger{}
)

func SetDefaultLogger(logger Logger) {
	if logger == nil {
		return
	}
	defaultLoggerMu.Lock()
	defer defaultLoggerMu.Unlock()
	defaultLogger = logger
}

func GetDefaultLogger() Logger {
	defaultLoggerMu.RLock()
	defer defaultLoggerMu.RUnlock()
	return defaultLogger
}

func formatFields(fields Fields) string {
	if len(fields) == 0 {
		return "{}"
	}
	return fmt.Sprintf("%v", map[string]any(fields))
}
