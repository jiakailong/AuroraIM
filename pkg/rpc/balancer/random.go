package balancer

import (
	"math/rand"
	"sync"
	"time"

	"kama_chat_server/pkg/rpc/registry"
)

type Random struct {
	mu  sync.Mutex
	rng *rand.Rand
}

func NewRandom() *Random {
	return &Random{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (random *Random) Pick(instances []registry.Instance, key string) (registry.Instance, error) {
	validInstances, err := validateInstances(instances)
	if err != nil {
		return registry.Instance{}, err
	}

	random.mu.Lock()
	index := random.rng.Intn(len(validInstances))
	random.mu.Unlock()

	return validInstances[index], nil
}
