package registry

import "fmt"

type StaticRegistry struct {
	services map[string][]Instance
}

func NewStaticRegistry(services map[string][]Instance) *StaticRegistry {
	registry := &StaticRegistry{services: make(map[string][]Instance, len(services))}
	for serviceName, instances := range services {
		registry.services[serviceName] = cloneInstances(instances)
	}
	return registry
}

func (registry *StaticRegistry) List(serviceName string) ([]Instance, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("registry: service name is empty")
	}
	instances, ok := registry.services[serviceName]
	if !ok {
		return nil, nil
	}
	return cloneInstances(instances), nil
}

func cloneInstances(instances []Instance) []Instance {
	if len(instances) == 0 {
		return nil
	}
	cloned := make([]Instance, len(instances))
	for index := range instances {
		cloned[index] = instances[index]
		if len(instances[index].Metadata) == 0 {
			continue
		}
		cloned[index].Metadata = make(map[string]string, len(instances[index].Metadata))
		for key, value := range instances[index].Metadata {
			cloned[index].Metadata[key] = value
		}
	}
	return cloned
}
