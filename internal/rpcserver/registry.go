package rpcserver

import (
	"fmt"
	"reflect"
	"strings"
)

var errorType = reflect.TypeOf((*error)(nil)).Elem()

func (s *Server) Register(name string, receiver any) error {
	// 1. Validate service name.
	name = strings.TrimSpace(name)

	if name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	if strings.Contains(name, ".") {
		return fmt.Errorf("service name cannot contain '.'")
	}

	// 2. Validate receiver.
	receiverValue := reflect.ValueOf(receiver)

	if !receiverValue.IsValid() {
		return fmt.Errorf("service receiver cannot be nil")
	}

	switch receiverValue.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map,
		reflect.Slice, reflect.Func, reflect.Chan:
		if receiverValue.IsNil() {
			return fmt.Errorf("service receiver cannot be nil")
		}
	}

	// 3. Create service metadata.
	svc := &Service{
		name:    name,
		methods: make(map[string]*MethodType),
	}

	receiverType := receiverValue.Type()

	// 4. Enumerate all exported methods.
	for i := 0; i < receiverType.NumMethod(); i++ {
		reflectedMethod := receiverType.Method(i)

		methodName := reflectedMethod.Name

		method := receiverValue.Method(i)
		signature := method.Type()

		if signature.IsVariadic() {
			continue
		}

		if signature.NumIn() != 1 {
			continue
		}

		if signature.NumOut() != 2 {
			continue
		}

		if signature.Out(1) != errorType {
			continue
		}

		argType := signature.In(0)
		replyType := signature.Out(0)

		svc.methods[methodName] = &MethodType{
			method:    method,
			argType:   argType,
			replyType: replyType,
		}

	}

	// 9. Reject services with no eligible methods.
	if len(svc.methods) == 0 {
		return fmt.Errorf(
			"service %q has no suitable methods",
			name,
		)
	}

	// 10. Register service atomically.
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.services[name]; exists {
		return fmt.Errorf("service %q is already registered", name)
	}

	s.services[name] = svc

	return nil
}
func (s *Server) lookupMethod(
	serviceName string,
	methodName string,
) (*MethodType, bool) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	svc, exists := s.services[serviceName]
	if !exists {
		return nil, false
	}

	method, exists := svc.methods[methodName]
	if !exists {
		return nil, false
	}

	return method, true
}
