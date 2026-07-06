package tcp

import (
	"context"
    "log/slog"
    "net"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct{
	address string
	handler Handler
	logger  *slog.Logger
	nextConnID atomic.Uint64
	wg sync.WaitGroup
	shutdownTimeout time.Duration
	readTimeout  time.Duration
    writeTimeout time.Duration
}

func NewServer(address string, handler Handler,logger *slog.Logger,shutdownTimeout time.Duration,readTimeout time.Duration,
    writeTimeout time.Duration,) *Server{
	return &Server{
		address: address,
		handler: handler,
		logger: logger,
		shutdownTimeout: shutdownTimeout,
		readTimeout:       readTimeout,
    	writeTimeout:      writeTimeout,
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
		if err != nil {
			if ctx.Err() != nil {
				s.logger.Info("waiting for active connections to finish")
				done := make(chan struct{})
				go func() {
					s.wg.Wait()
					close(done)
				}()
				select {

					case <-done:

						s.logger.Info("all active connections finished")

					case <-time.After(s.shutdownTimeout):

						s.logger.Warn(
							"shutdown timeout exceeded; forcing server shutdown",
							slog.Duration("timeout", s.shutdownTimeout),
						)
				}

				return nil
			}

			s.logger.Error(
				"failed to accept connection",
				slog.Any("error", err),
			)

			continue
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
		s.wg.Add(1)
		go func(){
			defer s.wg.Done()
			timeoutConn:=NewTimeoutConn(conn,s.readTimeout,s.writeTimeout)
			s.handler.Handle(timeoutConn, connLogger)
		}()
	}
}