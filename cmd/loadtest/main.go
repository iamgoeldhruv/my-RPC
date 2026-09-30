package main
import (
	"fmt"
	"io"
	"net"
)

const(
	serverAddress = "localhost:8080"
	messageCount  = 100
)

func runClient(address string, clientID int) error{
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()
	for i:=1;i<=messageCount;i++{
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

func main(){
	if err := runClient(serverAddress,1); err != nil {
		panic(err)
	}
	fmt.Println("client completed successfully")

}