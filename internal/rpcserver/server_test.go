package rpcserver

import (
    "encoding/json"
    "io"
    "log/slog"
    "net"
    "testing"

    "my-rpc/internal/framing"
    "my-rpc/internal/protocol"
)
func TestRPCServerHandle(t *testing.T) {
    server := NewServer()

    if err := server.Register("Calculator", &Calculator{}); err != nil {
        t.Fatal(err)
    }

    serverConn, clientConn := net.Pipe()

    defer clientConn.Close()

    logger := slog.New(
        slog.NewTextHandler(io.Discard, nil),
    )

    done := make(chan struct{})

    go func() {
        defer close(done)
        server.Handle(serverConn, logger)
    }()

    req := protocol.Request{
        ID:     1,
        Method: "Calculator.Add",
        Params: json.RawMessage(`{"a":10,"b":20}`),
    }

    requestPayload, err := protocol.EncodeRequest(&req)
    if err != nil {
        t.Fatal(err)
    }

    // Send request.
    if err := framing.WriteFrame(clientConn, requestPayload); err != nil {
        t.Fatal(err)
    }

    // Read response.
    responsePayload, err := framing.ReadFrame(clientConn)
    if err != nil {
        t.Fatal(err)
    }

    var response protocol.Response

    if err := json.Unmarshal(responsePayload, &response); err != nil {
        t.Fatal(err)
    }

    if response.ID != 1 {
        t.Fatalf(
            "expected response ID 1, got %d",
            response.ID,
        )
    }

    if response.Error != nil {
        t.Fatalf(
            "unexpected RPC error: %+v",
            response.Error,
        )
    }

    var result AddResponse

    if err := json.Unmarshal(response.Result, &result); err != nil {
        t.Fatal(err)
    }

    if result.Result != 30 {
        t.Fatalf(
            "expected result 30, got %d",
            result.Result,
        )
    }

    clientConn.Close()

    <-done
}

func TestRPCServerUnknownMethod(t *testing.T) {
    server := NewServer()

    if err := server.Register("Calculator", &Calculator{}); err != nil {
        t.Fatal(err)
    }

    logger := slog.New(
        slog.NewTextHandler(io.Discard, nil),
    )

    req := protocol.Request{
        ID:     2,
        Method: "Calculator.Unknown",
        Params: json.RawMessage(`{}`),
    }

    payload, err := protocol.EncodeRequest(&req)
    if err != nil {
        t.Fatal(err)
    }

    response, err := server.processRequest(payload, logger)
    if err != nil {
        t.Fatal(err)
    }

    if response.Error == nil {
        t.Fatal("expected method-not-found error")
    }

    if response.Error.Code != protocol.CodeMethodNotFound {
        t.Fatalf(
            "expected error code %d, got %d",
            protocol.CodeMethodNotFound,
            response.Error.Code,
        )
    }

    if response.ID != 2 {
        t.Fatalf("expected ID 2, got %d", response.ID)
    }
}
func TestRPCServerInvalidParameters(t *testing.T) {
    server := NewServer()

    if err := server.Register("Calculator", &Calculator{}); err != nil {
        t.Fatal(err)
    }

    logger := slog.New(
        slog.NewTextHandler(io.Discard, nil),
    )

    req := protocol.Request{
        ID:     3,
        Method: "Calculator.Add",
        Params: json.RawMessage(`{"a":"hello","b":20}`),
    }

    payload, err := protocol.EncodeRequest(&req)
    if err != nil {
        t.Fatal(err)
    }

    response, err := server.processRequest(payload, logger)
    if err != nil {
        t.Fatal(err)
    }

    if response.Error == nil {
        t.Fatal("expected invalid-parameters error")
    }

    if response.Error.Code != protocol.CodeInvalidRequest {
        t.Fatalf(
            "expected error code %d, got %d",
            protocol.CodeInvalidRequest,
            response.Error.Code,
        )
    }
}
func (c *Calculator) Fail(
    req AddRequest,
) (AddResponse, error) {
    return AddResponse{}, fmt.Errorf("calculation failed")
}
func TestRPCServerMethodError(t *testing.T) {
    server := NewServer()

    if err := server.Register("Calculator", &Calculator{}); err != nil {
        t.Fatal(err)
    }

    logger := slog.New(
        slog.NewTextHandler(io.Discard, nil),
    )

    req := protocol.Request{
        ID:     4,
        Method: "Calculator.Fail",
        Params: json.RawMessage(`{"a":10,"b":20}`),
    }

    payload, err := protocol.EncodeRequest(&req)
    if err != nil {
        t.Fatal(err)
    }

    response, err := server.processRequest(payload, logger)
    if err != nil {
        t.Fatal(err)
    }

    if response.Error == nil {
        t.Fatal("expected internal error")
    }

    if response.Error.Code != protocol.CodeInternalError {
        t.Fatalf(
            "expected code %d, got %d",
            protocol.CodeInternalError,
            response.Error.Code,
        )
    }
}