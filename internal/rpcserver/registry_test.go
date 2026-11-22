package rpcserver

import (
	"reflect"
	"testing"
)

type AddRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}

type AddResponse struct {
	Result int `json:"result"`
}

type Calculator struct{}

func (c *Calculator) Add(
	req AddRequest,
) (AddResponse, error) {
	return AddResponse{
		Result: req.A + req.B,
	}, nil
}

func (c *Calculator) Multiply(
	req AddRequest,
) (AddResponse, error) {
	return AddResponse{
		Result: req.A * req.B,
	}, nil
}

// Invalid: no arguments.
func (c *Calculator) Reset() {}

// Invalid: only one return value.
func (c *Calculator) Invalid(req AddRequest) AddResponse {
	return AddResponse{}
}

// Invalid: wrong second return type.
func (c *Calculator) AnotherInvalid(
	req AddRequest,
) (AddResponse, string) {
	return AddResponse{}, ""
}
func TestRegisterService(t *testing.T) {
	server := NewServer()

	err := server.Register("Calculator", &Calculator{})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	addMethod, ok := server.lookupMethod("Calculator", "Add")
	if !ok {
		t.Fatal("Add method was not registered")
	}

	if addMethod.argType != reflect.TypeOf(AddRequest{}) {
		t.Fatal("incorrect argument type")
	}

	if addMethod.replyType != reflect.TypeOf(AddResponse{}) {
		t.Fatal("incorrect reply type")
	}
}
func TestRegisterOnlySuitableMethods(t *testing.T) {
	server := NewServer()

	err := server.Register("Calculator", &Calculator{})
	if err != nil {
		t.Fatal(err)
	}

	invalidMethods := []string{
		"Reset",
		"Invalid",
		"AnotherInvalid",
	}

	for _, name := range invalidMethods {
		if _, ok := server.lookupMethod("Calculator", name); ok {
			t.Errorf("method %s should not be registered", name)
		}
	}
}
func TestRegisterDuplicateService(t *testing.T) {
	server := NewServer()

	if err := server.Register("Calculator", &Calculator{}); err != nil {
		t.Fatal(err)
	}

	err := server.Register("Calculator", &Calculator{})
	if err == nil {
		t.Fatal("expected error for duplicate service")
	}
}

type InvalidService struct{}

func (s *InvalidService) Reset() {}

func TestRegisterNoSuitableMethods(t *testing.T) {
	server := NewServer()

	err := server.Register("InvalidService", &InvalidService{})
	if err == nil {
		t.Fatal("expected error for service with no suitable methods")
	}
}
func TestRegisterNilReceiver(t *testing.T) {
	server := NewServer()

	var calculator *Calculator

	err := server.Register("Calculator", calculator)
	if err == nil {
		t.Fatal("expected error for nil receiver")
	}
}
