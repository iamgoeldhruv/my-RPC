package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "log"
    "my-rpc/internal/logger"
    "my-rpc/internal/transport/tcp"
    "time"
)



func main() {
    logger := logger.New()
    handler := tcp.NewEchoHandler()

    server := tcp.NewServer(
        tcp.Config{
            Address:":8080",
            ReadTimeout:30 * time.Second,
            WriteTimeout:30 * time.Second,
            ShutdownTimeout:30 * time.Second,

        },
       
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