package server

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrInvalidService           = errors.New("server: invalid service")
	ErrInvalidServiceMethodSign = errors.New("server: invalid service method signature")
)

var (
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
)

// RegisterService 通过反射注册服务方法，methodKey 格式为 Service.Method。
func (server *Server) RegisterService(service any) error {
	if service == nil {
		return ErrInvalidService
	}

	receiver := reflect.ValueOf(service)
	receiverType := receiver.Type()
	serviceType := receiverType
	if receiverType.Kind() == reflect.Pointer {
		serviceType = receiverType.Elem()
	}
	if serviceType.Kind() != reflect.Struct {
		return fmt.Errorf("%w: service must be struct or pointer to struct", ErrInvalidService)
	}

	serviceName := serviceType.Name()
	if serviceName == "" {
		return fmt.Errorf("%w: service name is empty", ErrInvalidService)
	}

	registered := 0
	for index := 0; index < receiverType.NumMethod(); index++ {
		method := receiverType.Method(index)
		if !method.IsExported() {
			continue
		}
		requestType, responseType, err := validateMethodSignature(method)
		if err != nil {
			continue
		}

		methodKey := fmt.Sprintf("%s.%s", serviceName, method.Name)
		handler, err := buildReflectedHandler(receiver, method, requestType, responseType)
		if err != nil {
			return err
		}
		server.Register(methodKey, handler)
		registered++
	}

	if registered == 0 {
		return fmt.Errorf("%w: no exported methods matched required signature", ErrInvalidServiceMethodSign)
	}

	return nil
}

func validateMethodSignature(method reflect.Method) (reflect.Type, reflect.Type, error) {
	methodType := method.Type
	if methodType.NumIn() != 3 || methodType.NumOut() != 2 {
		return nil, nil, ErrInvalidServiceMethodSign
	}

	if !methodType.In(1).Implements(contextType) {
		return nil, nil, ErrInvalidServiceMethodSign
	}

	requestType := methodType.In(2)
	if requestType.Kind() != reflect.Pointer || requestType.Elem().Kind() != reflect.Struct {
		return nil, nil, ErrInvalidServiceMethodSign
	}

	responseType := methodType.Out(0)
	if responseType.Kind() != reflect.Pointer || responseType.Elem().Kind() != reflect.Struct {
		return nil, nil, ErrInvalidServiceMethodSign
	}
	if !methodType.Out(1).Implements(errorType) {
		return nil, nil, ErrInvalidServiceMethodSign
	}

	return requestType, responseType, nil
}
