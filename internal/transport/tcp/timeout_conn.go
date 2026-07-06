package tcp
import(
"net"
"time"
)

type TimeoutConn struct {
	net.Conn

	readTimeout  time.Duration
	writeTimeout time.Duration
}

func (c *TimeoutConn) Read(b []byte) (int, error) {
	if err := c.Conn.SetReadDeadline(time.Now().Add(c.readTimeout)); err != nil {
		return 0, err
	}

	return c.Conn.Read(b)
}
func (c *TimeoutConn) Write(b []byte) (int, error) {
	if err := c.Conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); err != nil {
		return 0, err
	}

	return c.Conn.Write(b)
}

func NewTimeoutConn(
	conn net.Conn,
	readTimeout time.Duration,
	writeTimeout time.Duration,
) *TimeoutConn {
	return &TimeoutConn{
		Conn:         conn,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	}
}