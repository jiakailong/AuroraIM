package observability

import (
	"sync"
	"sync/atomic"
	"time"
)

type Metrics interface {
	IncInFlight(method string)
	DecInFlight(method string)
	Observe(method string, code uint16, latency time.Duration)
	Snapshot() MetricsSnapshot
}

type MetricsSnapshot struct {
	RequestsTotal map[string]uint64
	LatencyTotal  map[string]time.Duration
	InFlight      map[string]int64
}

type inMemoryMetrics struct {
	requestTotal sync.Map
	latencyNanos sync.Map
	inFlight     sync.Map
}

type atomicUint64 struct {
	value atomic.Uint64
}

type atomicInt64 struct {
	value atomic.Int64
}

func NewInMemoryMetrics() Metrics {
	return &inMemoryMetrics{}
}

func (metrics *inMemoryMetrics) IncInFlight(method string) {
	counter := metrics.getInt64Counter(&metrics.inFlight, method)
	counter.value.Add(1)
}

func (metrics *inMemoryMetrics) DecInFlight(method string) {
	counter := metrics.getInt64Counter(&metrics.inFlight, method)
	counter.value.Add(-1)
}

func (metrics *inMemoryMetrics) Observe(method string, code uint16, latency time.Duration) {
	key := method + "|" + codeKey(code)
	requestCounter := metrics.getUint64Counter(&metrics.requestTotal, key)
	requestCounter.value.Add(1)

	latencyCounter := metrics.getUint64Counter(&metrics.latencyNanos, key)
	latencyCounter.value.Add(uint64(latency.Nanoseconds()))
}

func (metrics *inMemoryMetrics) Snapshot() MetricsSnapshot {
	snapshot := MetricsSnapshot{
		RequestsTotal: make(map[string]uint64),
		LatencyTotal:  make(map[string]time.Duration),
		InFlight:      make(map[string]int64),
	}

	metrics.requestTotal.Range(func(key any, value any) bool {
		name, ok := key.(string)
		if !ok {
			return true
		}
		counter, ok := value.(*atomicUint64)
		if !ok {
			return true
		}
		snapshot.RequestsTotal[name] = counter.value.Load()
		return true
	})

	metrics.latencyNanos.Range(func(key any, value any) bool {
		name, ok := key.(string)
		if !ok {
			return true
		}
		counter, ok := value.(*atomicUint64)
		if !ok {
			return true
		}
		snapshot.LatencyTotal[name] = time.Duration(counter.value.Load())
		return true
	})

	metrics.inFlight.Range(func(key any, value any) bool {
		name, ok := key.(string)
		if !ok {
			return true
		}
		counter, ok := value.(*atomicInt64)
		if !ok {
			return true
		}
		snapshot.InFlight[name] = counter.value.Load()
		return true
	})

	return snapshot
}

func (metrics *inMemoryMetrics) getUint64Counter(storage *sync.Map, key string) *atomicUint64 {
	if existing, ok := storage.Load(key); ok {
		counter, ok := existing.(*atomicUint64)
		if ok {
			return counter
		}
	}
	counter := &atomicUint64{}
	actual, _ := storage.LoadOrStore(key, counter)
	return actual.(*atomicUint64)
}

func (metrics *inMemoryMetrics) getInt64Counter(storage *sync.Map, key string) *atomicInt64 {
	if existing, ok := storage.Load(key); ok {
		counter, ok := existing.(*atomicInt64)
		if ok {
			return counter
		}
	}
	counter := &atomicInt64{}
	actual, _ := storage.LoadOrStore(key, counter)
	return actual.(*atomicInt64)
}

func codeKey(code uint16) string {
	if code < 10 {
		return "00" + string(rune('0'+code))
	}
	if code < 100 {
		return "0" + itoa(code)
	}
	return itoa(code)
}

func itoa(value uint16) string {
	if value == 0 {
		return "0"
	}
	buffer := [5]byte{}
	index := len(buffer)
	for value > 0 {
		index--
		buffer[index] = byte('0' + value%10)
		value /= 10
	}
	return string(buffer[index:])
}

var (
	defaultMetricsMu sync.RWMutex
	defaultMetrics   Metrics = NewInMemoryMetrics()
)

func SetDefaultMetrics(metrics Metrics) {
	if metrics == nil {
		return
	}
	defaultMetricsMu.Lock()
	defer defaultMetricsMu.Unlock()
	defaultMetrics = metrics
}

func GetDefaultMetrics() Metrics {
	defaultMetricsMu.RLock()
	defer defaultMetricsMu.RUnlock()
	return defaultMetrics
}
