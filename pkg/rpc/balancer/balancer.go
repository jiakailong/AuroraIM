package balancer

import (
	"fmt"
	"strings"

	"kama_chat_server/pkg/rpc/registry"
)

type Balancer interface {
	Pick(instances []registry.Instance, key string) (registry.Instance, error)
}

func normalizeInstances(instances []registry.Instance) []registry.Instance {
	normalized := make([]registry.Instance, 0, len(instances))
	seen := make(map[string]struct{}, len(instances))
	for _, instance := range instances {
		address := strings.TrimSpace(instance.Address)
		if address == "" {
			continue
		}
		if _, ok := seen[address]; ok {
			continue
		}
		seen[address] = struct{}{}
		normalized = append(normalized, registry.Instance{
			Address:  address,
			Metadata: instance.Metadata,
		})
	}
	return normalized
}

func validateInstances(instances []registry.Instance) ([]registry.Instance, error) {
	normalized := normalizeInstances(instances)
	if len(normalized) == 0 {
		return nil, fmt.Errorf("balancer: no valid instance")
	}
	return normalized, nil
}
