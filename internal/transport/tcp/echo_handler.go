package tcp

import (
    "io"
    "log/slog"
    "net"
)

type EchoHandler struct{}

func NewEchoHandler() *EchoHandler{
	return &EchoHandler{}
}

func (h *EchoHandler) Handle(conn net.Conn, logger *slog.Logger) {
	defer conn.Close()
	logger.Info("handling connection")
	buffer:=make([]byte, 4096)
	for{
		n,err:=conn.Read(buffer)
		if err != nil {
            if err != io.EOF {
                logger.Error(
					"failed to read from connection",
					slog.Any("error", err),
				)
            }
            return
        }


	
		_,err=conn.Write(buffer[:n])
		if err != nil {
			logger.Error(
				"failed to write to connection",
				slog.Any("error", err),
			)
			return
		}
		
	}
}