package main
import (
	"fmt"
	"io"
	"net"
	"time"
	"sync"
)

const(
	serverAddress = "localhost:8080"
	messageCount  = 100
	clientTimeout = 5 * time.Second
	clientCount=1000
)

func runClient(address string, clientID int) error{
	conn, err := net.DialTimeout("tcp", address,clientTimeout)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()
	for i:=1;i<=messageCount;i++{
		if err := conn.SetDeadline(time.Now().Add(clientTimeout)); err != nil {
        return fmt.Errorf("set deadline: %w", err)
    }
		expected := fmt.Sprintf("client-%d-message-%d", clientID, i)
		message := []byte(expected)

		if _, err := conn.Write(message); err != nil {
			return fmt.Errorf("write message %d: %w", i, err)
		}
		response := make([]byte, len(message))
		if _, err := io.ReadFull(conn, response); err != nil {
			return fmt.Errorf("read response %d: %w", i, err)
		}
		if string(response)!=expected{
			return fmt.Errorf(
				"invalid response for message %d: got %q, expected %q",
				i,
				string(response),
				string(message),
			)
		}

	}
	return nil

}

func main() {
    var wg sync.WaitGroup

    for clientID := 1; clientID <= clientCount; clientID++ {
        wg.Add(1)

        go func(id int) {
            defer wg.Done()

            if err := runClient(serverAddress, id); err != nil {
                fmt.Printf("client %d failed: %v\n", id, err)
            }
        }(clientID)
    }

    wg.Wait()

    fmt.Println("load test completed")
}