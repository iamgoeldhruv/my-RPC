package rpcserver

import (
	"reflect"
	"sync"
)

type MethodType struct {
	method    reflect.Value
	argType   reflect.Type
	replyType reflect.Type
}

type Service struct {
	name    string
	methods map[string]*MethodType
}
type Server struct {
	mu       sync.RWMutex
	services map[string]*Service
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]*Service),
	}
}
