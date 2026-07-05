package main

import (
    "log"
    "my-rpc/internal/logger"
    "my-rpc/internal/transport/tcp"
)

func main() {
    logger := logger.New()
    handler := tcp.NewEchoHandler()

    server := tcp.NewServer(
        ":8080",
        handler,
        logger,
    )

    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}