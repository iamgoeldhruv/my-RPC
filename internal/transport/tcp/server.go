package tcp

import (
	"context"
    "log/slog"
    "net"
	"sync/atomic"
)

type Server struct{
	address string
	handler Handler
	logger  *slog.Logger
	nextConnID atomic.Uint64
}

func NewServer(address string, handler Handler,logger *slog.Logger,) *Server{
	return &Server{
		address: address,
		handler: handler,
		logger: logger,
	}
}

func (s *Server) Run(ctx context.Context) error{
	listener,err:=net.Listen("tcp",s.address)
		go func() {

		<-ctx.Done()

		s.logger.Info("shutdown signal received")

		listener.Close()

	}()
	if err!=nil{
		return err
	}
	s.logger.Info(
		"TCP server listening",
		slog.String("address", s.address),
	)
	for{
		conn,err:=listener.Accept()
		if err!=nil{
			if ctx.Err() != nil {
				s.logger.Info("TCP server stopped gracefully")
				return nil
			}

			s.logger.Error(
				"failed to accept connection",
				slog.Any("error", err),
			)
			
		}

		s.logger.Info(
			"connection accepted",
			slog.String("remote_addr", conn.RemoteAddr().String()),
		)
		connID := s.nextConnID.Add(1)
		connLogger := s.logger.With(
			slog.Uint64("conn_id", connID),
			slog.String("remote_addr", conn.RemoteAddr().String()),
		)
		go s.handler.Handle(conn, connLogger)
	}
}