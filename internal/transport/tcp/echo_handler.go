package tcp

import (
	"errors"
	"io"
	"log/slog"
	"my-rpc/internal/framing"
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

	for {
		payload, err := framing.ReadFrame(conn)

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

		err = framing.WriteFrame(conn, payload)
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
