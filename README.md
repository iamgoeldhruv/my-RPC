# My-RPC

A from-scratch Remote Procedure Call (RPC) framework implemented in Go.

The goal of this project is to build an RPC framework layer by layer, starting with the transport layer and gradually adding message framing, serialization, service registration, request dispatching, and client/server abstractions. Each component is implemented from first principles to understand how production RPC systems are designed.

---

## Current Features

The project currently provides a reusable TCP transport layer with the following capabilities:

- TCP server implementation
- Centralized TCP server configuration through a Config structure
- Concurrent connection handling (goroutine per connection)
- Pluggable connection handlers
- Echo handler for transport validation
- Structured logging using Go's `log/slog`
- Per-connection logging with unique connection IDs
- Graceful server shutdown using `context.Context`
- Signal-based shutdown (`SIGINT`/`SIGTERM`)
- Active connection tracking using `sync.WaitGroup`
- Graceful draining of in-flight connections during shutdown
- Configurable shutdown timeout to prevent indefinite blocking
- Configurable per-connection read timeouts
- Configurable per-connection write timeouts
- Automatic deadline refresh before every read/write operation
- Graceful handling of idle and stalled client connections

At this stage, the project focuses solely on the networking infrastructure. No RPC protocol has been implemented yet.

---

## Project Structure

```text
.
├── cmd
│   └── server
│       └── main.go
│
├── internal
│   ├── logger
│   │   └── logger.go
│   │
│   └── transport
│       └── tcp
│           ├── handler.go
│           ├── echo_handler.go
|           └── timeout_conn.go
│           └── server.go
│
├── go.mod
└── README.md
```

### `cmd/server`

Application entry point.

Responsible for:

- Creating the application logger
- Creating the TCP server configuration
- Initializing the TCP server
- Wiring dependencies
- Creating the root application context
- Handling operating system shutdown signals (`SIGINT`/`SIGTERM`)
- Starting the server

## Server Configuration

TCP server configuration is encapsulated in a `Config` structure.

```go
type Config struct {
    Address         string
    ReadTimeout     time.Duration
    WriteTimeout    time.Duration
    ShutdownTimeout time.Duration
}

### `internal/logger`

Provides the application's structured logging configuration.

Uses Go's standard `log/slog` package and exposes a single logger that is injected into the server. The server creates child loggers for each accepted connection, automatically attaching contextual information such as:

- Connection ID
- Remote client address

This keeps logging consistent while avoiding repeated log fields throughout the codebase.

### `internal/transport/tcp`

Implements the networking layer.

The transport layer is intentionally independent of any RPC-specific logic.

Responsibilities include:

- Listening on a TCP address
- Accepting incoming connections
- Assigning unique connection IDs
- Creating connection-scoped loggers
- Spawning a goroutine for every client
- Delegating connection processing to a handler
- Tracking active client connections
- Reacting to context cancellation
- Stopping acceptance of new connections during shutdown
- Waiting for active connections to complete
- Forcing shutdown after a configurable timeout
- Applying configurable read/write deadlines to every connection
- Detecting idle or stalled clients through connection deadlines
- Enforcing connection-level timeout policies independently of protocol logic

---

## Architecture

```text
                         Application
                              │
                              │
                   signal.NotifyContext()
                              │
                              ▼
                     context.Context
                              │
                              ▼
                     +----------------+
                     |   TCP Server   |
                     +----------------+
                              │
               ┌──────────────┴────────────────────────────────────┐
               │                                                   │
               ▼                                                   ▼
      Accept TCP Connections                              Wait for Context
               │                                                   │
               ▼                                                   ▼
      Generate Connection ID                            Shutdown Signal
               │                                                   │
               ▼                                                   ▼
      Create Child Logger                               listener.Close()
               │                                                   │
               ▼                                                   ▼
      Wrap Connection with                           Stop Accepting New Clients
          TimeoutConn                                           │
               │                                                ▼
               ▼                                   Wait for Active Connections
      WaitGroup.Add(1)                                         │
               │                                   ┌────────────┴────────────┐
               ▼                                   ▼                         ▼
        Handler Interface                All Connections Done      Shutdown Timeout
               │                                   │                         │
               ▼                                   └────────────┬────────────┘
          EchoHandler                                         ▼
               │                                         Server Exit
               │
        ┌──────┴──────┐
        ▼             ▼
   Read()         Write()
        │             │
        ▼             ▼
SetReadDeadline  SetWriteDeadline
        │             │
        └──────┬──────┘
               ▼
      Underlying net.Conn
               │
               ▼
     Timeout/Error Returned
               │
               ▼
    Close Connection Gracefully
                       
```

The transport layer manages networking concerns only.

Application behavior is delegated through the `Handler` interface, allowing different protocols to be implemented without modifying the transport layer.

---

## Handler Abstraction

The transport layer communicates only through the following interface:

```go
type Handler interface {
    Handle(net.Conn, *slog.Logger)
}
```

Every accepted connection receives:

- The network connection
- A connection-scoped logger

This keeps the transport layer reusable while enabling handlers to produce structured logs without knowing how logging is configured.

---

## Structured Logging

Logging is implemented using Go's standard `log/slog` package.

A single application logger is created during startup and injected into the TCP server.

For every accepted connection, the server creates a child logger containing connection-specific context.

Example log output:

```text
time=2026-07-06T12:30:11Z
level=INFO
msg="connection accepted"
conn_id=3
remote_addr=127.0.0.1:51243
```

This design allows future RPC request logs to automatically include additional context such as request IDs, service names, and method names.

---

## Running

Start the server:

```bash
go run ./cmd/server
```

The server listens on:

```text
:8080
```

---

## Testing

Using `netcat`:

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

The current implementation simply echoes every received message back to the client.

---

## Design Principles

This project follows a few guiding principles:

- Keep networking independent of protocol implementation.
- Separate application lifecycle management from transport responsibilities.
- Prefer dependency injection over global state.
- Build reusable components with clear responsibilities.
- Use structured logging from the beginning.
- Design components around `context.Context` for cancellation and graceful lifecycle management.
- Gracefully drain in-flight work before process termination.
- Evolve the framework incrementally while maintaining a clean architecture.
- Keep connection-level policies (timeouts, lifecycle management) inside the transport layer.