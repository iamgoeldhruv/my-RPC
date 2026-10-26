package tcp

import (
	"bytes"
	"io"
	"log/slog"
	"net"
	"testing"

	"my-rpc/internal/framing"
)

func TestEchoHandlerFramedMessages(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := &EchoHandler{}

	handlerDone := make(chan struct{})
	go func() {
		defer close(handlerDone)
		handler.Handle(serverConn, logger)
	}()

	payloads := [][]byte{
		[]byte("hello"),
		{},
		{0x00, 0xFF, 0x10, 0x80},
		[]byte("second frame on the same connection"),
	}

	for i, want := range payloads {
		if err := framing.WriteFrame(clientConn, want); err != nil {
			t.Fatalf("write frame %d: %v", i, err)
		}

		got, err := framing.ReadFrame(clientConn)
		if err != nil {
			t.Fatalf("read frame %d: %v", i, err)
		}

		if !bytes.Equal(got, want) {
			t.Fatalf("frame %d: got %x, want %x", i, got, want)
		}
	}

	if err := clientConn.Close(); err != nil {
		t.Fatalf("close client connection: %v", err)
	}

	<-handlerDone
}
