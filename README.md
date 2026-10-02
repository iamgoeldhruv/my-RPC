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
- Concurrent TCP load-test client
- 1000 concurrent client connections
- Persistent TCP connection testing
- 100 sequential request/response cycles per client
- Response validation for every request
- Synchronization of concurrent clients using `sync.WaitGroup`
- Structured per-message test results
- Per-message latency measurement
- Total test duration measurement
- Throughput measurement in messages per second
- Average latency calculation
- P50, P95, P99, and maximum latency calculation
- Client-side connection timeout using `net.DialTimeout`
- Client-side I/O deadlines to prevent load tests from hanging indefinitely

At this stage, the project focuses on the networking infrastructure and transport validation. No RPC protocol has been implemented yet.


---

## Project Structure

```text
.
├── cmd
│   ├── server
│   │   └── main.go
│   │
│   └── loadtest
│       ├── main.go
│       ├── client.go
│       ├── metrics.go
│       └── report.go
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
### `cmd/loadtest`

The load-test client is split into focused components:

- `main.go` — orchestrates concurrent clients and result collection
- `client.go` — manages TCP connections, request/response validation, and per-message results
- `metrics.go` — calculates throughput and latency statistics
- `report.go` — formats and prints load-test results

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
### Concurrent TCP Load Test

The load-test client establishes 1000 concurrent persistent TCP connections.

Each client runs independently in its own goroutine and performs 100 sequential request/response cycles.

The load test uses sync.WaitGroup to wait for all client goroutines to complete.

Each client follows:

Connect
   ↓
Send 100 uniquely identified messages
   ↓
Validate every response
   ↓
Close connection
   ↓
Wait for all clients

Each client performs:

100 writes

100 reads

100 response validations

Across 1000 concurrent clients, the load test performs:

100,000 writes

100,000 reads

100,000 response validations
Each request/response cycle records its latency in a structured test result.

After all clients complete, the load test aggregates the results and reports:

- Total messages
- Successful messages
- Failed messages
- Total test duration
- Message throughput
- Average latency
- P50 latency
- P95 latency
- P99 latency
- Maximum latency
Example output:

Load Test Results
=================

Connections:       1000
Messages:          100000
Successful:        100000
Failed:            0

Duration:          2.41s
Throughput:        41493.78 msg/s

Average latency:   1.82ms
P50:               1.51ms
P95:               3.20ms
P99:               7.42ms
Max:               15.31ms
The load-test client uses a 5-second timeout for connection establishment and
refreshes the connection deadline before each request/response cycle.

This ensures that a stalled server or connection causes the test to fail
instead of hanging indefinitely.
Each request/response cycle also refreshes the connection deadline using
`conn.SetDeadline`, ensuring that individual network operations cannot block
indefinitely.

The deadline is refreshed before every message, so the 5-second timeout
applies independently to each request/response cycle.

Run the load test with:

```bash
go run ./cmd/loadtest

The command prints total messages, successes, failures, duration, throughput,
average latency, and P50/P95/P99/max latency. It exits with a non-zero status
when any message fails, which makes it suitable for CI or a repeatable local
performance check.

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