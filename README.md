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
- Length-prefix frame encoder using a 4-byte big-endian payload length
- Configurable protocol constant for the 16 MiB maximum payload size
- Payload size validation before writing
- Complete frame writes with partial-write handling
- Propagation of writer errors during frame encoding
- Unit tests for normal, empty, binary, oversized payloads, partial writes, writer failures, and maximum-size payloads
- Length-prefix frame decoder using `ReadFrame(io.Reader) ([]byte, error)`
- Complete header and payload reads using `io.ReadFull`
- Payload length validation before memory allocation
- Handling of truncated headers and payloads
- Support for empty and binary payloads during frame decoding
- Unit tests for valid frames, partial reads across multiple Read calls, truncated length headers and payloads, oversized declared lengths, and     malformed-input error handling
- Framed echo handler using `ReadFrame` and `WriteFrame`
- Length-prefixed request and response messages over persistent TCP connections
- Support for arbitrary binary payloads in the echo handler
- Load-test client using framed requests and responses
- Integration tests for binary payloads and multiple frames per connection
- Go fuzz testing for `ReadFrame` using arbitrary byte sequences
- Fuzz validation of malformed headers, oversized lengths,
  truncated payloads, and successful payload decoding
- Benchmarks for `WriteFrame` and `ReadFrame` using Go's testing package
- Benchmark allocation reporting with `-benchmem`

At this stage, the project includes a reusable TCP transport layer, a length-prefixed framing protocol, and an independent JSON-based RPC message protocol.

The JSON protocol defines request and response structures, request identifiers, method names, parameters, success results, and structured errors. It includes JSON encoding, decoding, message validation, and unit tests for valid messages and malformed inputs.

The framing layer remains responsible for message boundaries, while the protocol layer defines the structure and semantics of RPC messages.

The JSON protocol has not yet been integrated into the TCP handler. Service registration, request dispatching, and client/server RPC abstractions remain future work.


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
├── docs
│   └── framing
│       └── README.md
├── internal
│   ├── framing
│   │   ├── frame.go
│   │   ├── frame_test.go
│   │   └── frame_bench_test.go
|   ├── protocol
|   │   ├── message.go
|   │   ├── message_test.go
|   │   └── errors.go
│   ├── logger
│   │   └── logger.go
│   └── transport
│       └── tcp
│           ├── handler.go
│           ├── echo_handler.go
│           ├── timeout_conn.go
│           └── server.go
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
### `internal/framing`

Implements the message framing layer independently of the TCP transport.

Current responsibilities:

- Encoding payload lengths as unsigned 32-bit big-endian integers
- Writing length-prefixed frames using `WriteFrame(io.Writer, []byte) error`
- Reading length-prefixed frames using `ReadFrame(io.Reader) ([]byte, error)`
- Enforcing the maximum payload size of 16 MiB during encoding and decoding
- Handling partial writes using a full-write helper
- Handling partial header and payload reads using `io.ReadFull`
- Validating payload lengths before memory allocation
- Propagating contextual errors for header and payload I/O failures
- Unit testing valid frames, empty and binary payloads, partial reads, truncated headers and payloads, oversized declared lengths, and malformed-input handling
- Processing length-prefixed frames through the pluggable handler
- Keeping framing and message-processing logic outside the TCP server lifecycle implementation
- Fuzz testing `ReadFrame` against arbitrary and malformed byte sequences
  to validate decoder robustness and payload correctness
- Benchmarking frame encoding and decoding across different payload sizes
- Measuring execution time and memory allocations with Go benchmarks

The framing package does not manage connections, deadlines, or connection lifecycle.


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
### `internal/protocol`

Implements the JSON-based RPC message protocol independently of the transport and framing layers.

Current responsibilities:

- Defining RPC request structures with JSON tags.
- Defining RPC response structures for successful and failed calls.
- Representing request identifiers using unsigned 64-bit integers.
- Representing fully qualified method names using the `Service.Method` format.
- Using `json.RawMessage` for method parameters and response results to support different argument and return types.
- Defining structured RPC errors with numeric error codes and descriptive messages.
- Encoding RPC requests and responses into JSON.
- Decoding incoming JSON requests into typed request structures.
- Validating request method names, parameters, and response invariants.
- Ensuring responses contain either a result or an error, but not both.
- Providing helpers for constructing successful and error responses.
- Unit testing JSON serialization, deserialization, validation, and malformed messages.

The protocol package does not manage TCP connections, message framing, or service method execution.

It operates on JSON payloads and remains independent of the transport layer and framing implementation.
## JSON RPC Wire Protocol

The RPC protocol uses JSON to represent requests and responses. Each JSON message is transported as a length-prefixed frame using the existing framing layer.

### Request Format

Every RPC request contains:

- `id`: Unique request identifier used to correlate responses.
- `method`: Fully qualified method name in the `Service.Method` format.
- `params`: JSON object containing method arguments.

Example:

```json
{
  "id": 1,
  "method": "Calculator.Add",
  "params": {
    "a": 2,
    "b": 3
  }
}
```

### Successful Response

A successful response echoes the request identifier and contains the method result.

```json
{
  "id": 1,
  "result": 5
}
```

### Error Response

A failed request returns a structured error containing a numeric code and descriptive message.

```json
{
  "id": 1,
  "error": {
    "code": 404,
    "message": "method not found"
  }
}
```

### Initial Error Codes

| Code | Meaning |
| --- | --- |
| 400 | Invalid request |
| 404 | Method not found |
| 500 | Internal server error |

### Protocol Design

- Request identifiers allow clients to correlate responses with their original requests.
- `json.RawMessage` keeps method parameters and results independent of specific Go types.
- Request and response validation enforces protocol-level invariants.
- Successful responses contain a result, while failed responses contain a structured error.
- JSON serialization is kept separate from message framing and TCP transport.

The protocol package currently defines and validates these messages independently. Integration with the TCP handler and RPC dispatcher will be implemented in subsequent stages.
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

The echo handler uses the framing package to read complete messages with
`ReadFrame` and write responses with `WriteFrame`. It supports multiple
frames on a single connection and echoes payload bytes without interpreting
them as text.

The transport layer continues to own connection lifecycle management,
while the handler owns echo behavior.

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

### Manual Framed Echo Test

Start the server:

```bash
go run ./cmd/server
```

Send a length-prefixed message using netcat:

```bash
printf '\x00\x00\x00\x05hello' | nc -N localhost 8080 | xxd
```

Expected output:

```text
00000000: 0000 0005 6865 6c6c 6f
```

The first four bytes encode the payload length as a big-endian unsigned
integer. The remaining bytes contain the echoed payload.

Test binary data:

```bash
printf '\x00\x00\x00\x04\x00\xFF\x01\xFE' | nc -N localhost 8080 | xxd
```

Test multiple frames on one connection:

```bash
printf '\x00\x00\x00\x05hello\x00\x00\x00\x03rpc' | nc -N localhost 8080 | xxd
```

`nc -N` is supported by OpenBSD netcat. Depending on your installed
netcat implementation, you may need to use `nc -q 0` instead.


The echo handler reads and writes length-prefixed frames. Each request is
decoded into a payload and echoed back in a response frame. Arbitrary binary
payloads and multiple frames over a persistent TCP connection are supported.

### Framing Unit Tests

Run the framing package tests:

```bash
go test ./internal/framing
---
### Scenario-Based TCP Load Test

The load-test command runs an end-to-end scenario against a server process
that it starts automatically. Each scenario opens up to 1000 concurrent
persistent TCP connections, sends length-prefixed request frames, reads
length-prefixed response frames, validates echoed payloads, records
per-message latency, and prints aggregate metrics.

Run the default scenario:

```bash
go run ./cmd/loadtest
```

Select a scenario with the `--scenario` flag:

```bash
go run ./cmd/loadtest --scenario=normal
go run ./cmd/loadtest --scenario=before-start
go run ./cmd/loadtest --scenario=halfway
go run ./cmd/loadtest --scenario=mostly-done
go run ./cmd/loadtest --scenario=idle
```

Available scenarios:

| Scenario | Behavior |
| --- | --- |
| `normal` | Every client sends and validates all 100 messages, then the test completes normally. |
| `before-start` | All clients connect first; the server is shut down before message processing begins. |
| `halfway` | The server is shut down after every client reaches message 50. |
| `mostly-done` | The server is shut down after every client reaches message 90. |
| `idle` | 1000 idle connections are opened, held briefly, and then the server is shut down. |

The normal, halfway, and mostly-done scenarios perform 100 sequential
request/response cycles per client. With the default 1000 clients, that is up
to 100,000 writes, reads, and response validations. The client uses a 5-second
connection and per-request I/O timeout.

Each completed run reports:

- Connections
- Total messages
- Successful and failed messages
- Duration
- Message throughput
- Average latency
- P50, P95, and P99 latency
- Maximum latency
- Server shutdown errors, when applicable

Example report:

```text
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
```

The scenario runner is intended for local transport and graceful-shutdown
validation. Run one scenario at a time because each run starts a server on
`localhost:8080` and manages its shutdown lifecycle.

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

### Framing Fuzz Tests

The framing package includes Go fuzz tests for `ReadFrame` to validate
decoder robustness against arbitrary byte sequences and malformed input.

The fuzz test checks:

- Incomplete length headers
- Oversized declared payload lengths
- Truncated payloads
- Correct decoding of valid frames
- Preservation of arbitrary binary payloads
- Error handling for malformed input

Run the fuzz test:

```bash
go test -fuzz=FuzzReadFrame -fuzztime=5m ./internal/framing
```
### Protocol Unit Tests

The protocol package includes unit tests for JSON request and response serialization, deserialization, and validation.

The tests cover:

- Decoding valid RPC requests.
- Encoding RPC requests.
- Creating and encoding successful responses.
- Creating and encoding structured error responses.
- Validating request method names and parameters.
- Rejecting missing request identifiers.
- Rejecting invalid JSON parameters.
- Ensuring successful responses do not omit zero-valued results.
- Ensuring responses contain exactly one of `result` or `error`.
- Handling `nil` results as JSON `null`.

Run the protocol package tests:

```bash
go test ./internal/protocol
```

Run all project tests:

```bash
go test ./...
```

Run tests with the race detector:

```bash
go test -race ./...
```

Go automatically mutates seed inputs to explore additional input combinations.
Failing inputs are saved in the fuzz corpus for reproducibility and regression testing.

### Framing Benchmarks

The framing package includes benchmarks for `WriteFrame` and `ReadFrame` across
empty, small (64 B), medium (1 KB), and large (64 KB) payloads.

The benchmarks report execution time, throughput, bytes allocated per operation,
and allocations per operation. The read benchmark prepares the encoded frame
outside the timed section so it measures decoding rather than frame setup.

Run all framing benchmarks with allocation statistics:

```bash
go test -run=^$ -bench=. -benchmem ./internal/framing

## Documentation

- [Length-Prefix Framing Protocol — Design](docs/framing/README.md)

The framing documentation defines the four-byte big-endian length prefix,
the 16 MiB maximum payload, zero-length payload behavior, oversized-frame
rejection, and partial-read handling.

The `WriteFrame` encoder and `ReadFrame` decoder have been implemented and
integrated with the framed echo handler. End-to-end integration tests and
load-test scenarios validate framed request/response communication over TCP.
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