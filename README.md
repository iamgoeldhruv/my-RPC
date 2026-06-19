# My-RPC

A from-scratch RPC framework implementation in Go.

The project is built incrementally, starting from the transport layer and progressing toward a fully functional RPC system with request framing, serialization, method registration, dispatching, and client/server abstractions.

---

## Current Status

### Day 1: TCP Echo Server

Implemented a modular TCP server capable of:

- Listening for incoming TCP connections
- Handling multiple concurrent clients
- Reading raw bytes from connections
- Echoing received data back to clients
- Decoupling transport and application logic through a handler abstraction

This serves as the networking foundation for future RPC functionality.

---

## Project Structure

```text
.
├── cmd
│   └── server
│       └── main.go
│
├── internal
│   └── transport
│       └── tcp
│           ├── handler.go
│           ├── echo_handler.go
│           └── server.go
│
└── README.md
```

---

## Architecture

```text
┌─────────┐
│ Client  │
└────┬────┘
     │ TCP
     ▼
┌─────────┐
│ Server  │
└────┬────┘
     │
     ▼
┌─────────┐
│ Handler │
└────┬────┘
     │
     ▼
┌─────────────┐
│ EchoHandler │
└─────────────┘
```

The transport layer is responsible only for connection management.

Application behavior is delegated to a handler implementation.

This separation ensures the networking layer remains unchanged as the project evolves from a simple echo server into a complete RPC framework.

---

## Components

### TCP Server

Responsible for:

- Binding to a TCP address
- Accepting incoming connections
- Spawning a goroutine per connection
- Delegating connection processing to a handler

### Handler Interface

Defines the contract for processing a connection.

```go
type Handler interface {
    Handle(net.Conn)
}
```

Any component implementing this interface can be attached to the TCP server.

### Echo Handler

Current handler implementation.

For each connection:

1. Read bytes from the socket
2. Write the same bytes back to the client

---

## Running

Start the server:

```bash
go run ./cmd/server
```

Expected output:

```text
TCP server listening on :8080
```

---

## Testing

Using netcat:

```bash
nc localhost 8080
```

Example session:

```text
hello
hello

rpc
rpc
```

The server echoes every received message back to the client.

---

## Concurrency Model

Each accepted connection is handled in an independent goroutine.

```go
for {
    conn, err := listener.Accept()
    if err != nil {
        continue
    }

    go handler.Handle(conn)
}
```

This enables multiple clients to communicate with the server concurrently without blocking each other.

---

## Design Goals

- Keep transport concerns isolated from business logic
- Avoid coupling networking code to protocol implementation
- Build reusable primitives for future RPC features
- Maintain a clean architecture that can evolve incrementally

---

## Roadmap

### Day 1

- [x] TCP listener
- [x] Connection accept loop
- [x] Goroutine-per-connection model
- [x] Echo handler
- [x] Transport abstraction

