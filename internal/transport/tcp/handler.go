package tcp
import ("net"
"log/slog")

type Handler interface {
	Handle(conn net.Conn, logger *slog.Logger)
}