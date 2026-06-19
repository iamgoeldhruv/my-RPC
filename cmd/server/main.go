package main

import (
    "log"

    "my-rpc/internal/transport/tcp"
)

func main() {
    handler := tcp.NewEchoHandler()

    server := tcp.NewServer(
        ":8080",
        handler,
    )

    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}