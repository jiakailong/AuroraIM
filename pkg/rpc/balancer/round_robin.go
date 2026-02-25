package balancer

import (
	"sync/atomic"

	"kama_chat_server/pkg/rpc/registry"
)

type RoundRobin struct {
	cursor atomic.Uint64
}

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

func (roundRobin *RoundRobin) Pick(instances []registry.Instance, key string) (registry.Instance, error) {
	validInstances, err := validateInstances(instances)
	if err != nil {
		return registry.Instance{}, err
	}

	index := roundRobin.cursor.Add(1) - 1
	return validInstances[index%uint64(len(validInstances))], nil
}
