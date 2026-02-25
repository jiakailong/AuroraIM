package registry

import (
	"fmt"
	"sync"
)

type MemoryRegistry struct {
	mu       sync.RWMutex
	services map[string][]Instance
}

func NewMemoryRegistry() *MemoryRegistry {
	return &MemoryRegistry{services: make(map[string][]Instance)}
}

func (registry *MemoryRegistry) List(serviceName string) ([]Instance, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("registry: service name is empty")
	}

	registry.mu.RLock()
	defer registry.mu.RUnlock()
	instances, ok := registry.services[serviceName]
	if !ok {
		return nil, nil
	}
	return cloneInstances(instances), nil
}

func (registry *MemoryRegistry) Set(serviceName string, instances []Instance) error {
	if serviceName == "" {
		return fmt.Errorf("registry: service name is empty")
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.services[serviceName] = cloneInstances(instances)
	return nil
}
