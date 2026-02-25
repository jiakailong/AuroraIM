package registry

type Instance struct {
	Address  string
	Metadata map[string]string
}

type Registry interface {
	List(serviceName string) ([]Instance, error)
}
