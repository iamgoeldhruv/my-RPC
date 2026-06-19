package tcp

import (
    "log"
    "net"
)

type Server struct{
	address string
	handler Handler
}

func NewServer(address string, handler Handler) *Server{
	return &Server{
		address: address,
		handler: handler,
	}
}

func (s *Server) Start() error{
	listener,err:=net.Listen("tcp",s.address)
	if err!=nil{
		return err
	}
	log.Printf("TCP server listening on %s", s.address)
	for{
		conn,err:=listener.Accept()
		if err!=nil{
			log.Printf("accept error: %v", err)
			continue
		}
		go s.handler.Handle(conn)
	}
}