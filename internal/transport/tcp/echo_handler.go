package tcp

import (
    "io"
    "log"
    "net"
)

type EchoHandler struct{}

func NewEchoHandler() *EchoHandler{
	return &EchoHandler{}
}

func (h *EchoHandler) Handle(conn net.Conn){
	defer conn.Close()
	buffer:=make([]byte, 4096)
	for{
		n,err:=conn.Read(buffer)
		if err != nil {
            if err != io.EOF {
                log.Printf("read error: %v", err)
            }
            return
        }


	
		_,err=conn.Write(buffer[:n])
		if err != nil {
			log.Printf("write error: %v", err)
			return
		}
		
	}
}