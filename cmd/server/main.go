package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
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
    ctx, stop := signal.NotifyContext(
        context.Background(),
        os.Interrupt,
        syscall.SIGTERM,
    )
    defer stop()

    if err := server.Run(ctx); err != nil {
        log.Fatal(err)
    }
}