package tcp

import (
	"errors"
	"io"
	"log/slog"
	"net"
)

type EchoHandler struct{}

func NewEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

func isTimeoutError(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func (h *EchoHandler) Handle(conn net.Conn, logger *slog.Logger) {
	defer conn.Close()

	logger.Info("handling connection")

	buffer := make([]byte, 4096)

	for {
		n, err := conn.Read(buffer)
		if err != nil {

			if err == io.EOF {
				logger.Info("client disconnected")
				return
			}

			if isTimeoutError(err) {
				logger.Info(
					"connection timed out",
					slog.Any("error", err),
				)
				return
			}

			logger.Error(
				"failed to read from connection",
				slog.Any("error", err),
			)
			return
		}

		_, err = conn.Write(buffer[:n])
		if err != nil {

			if isTimeoutError(err) {
				logger.Info(
					"write timed out",
					slog.Any("error", err),
				)
				return
			}

			logger.Error(
				"failed to write to connection",
				slog.Any("error", err),
			)
			return
		}
	}
}