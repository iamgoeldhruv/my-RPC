package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

type Result struct {
	ClientID  int
	MessageID int
	Latency   time.Duration
	Error     error
}

func runClient(address string, clientID int) []Result {
	results := make([]Result, 0, messageCount)

	conn, err := net.DialTimeout("tcp", address, clientTimeout)
	if err != nil {
		results = append(results, Result{
			ClientID: clientID,
			Error:    fmt.Errorf("connect: %w", err),
		})

		return results
	}

	defer conn.Close()

	for messageID := 1; messageID <= messageCount; messageID++ {
		startTime := time.Now()

		if err := conn.SetDeadline(time.Now().Add(clientTimeout)); err != nil {
			results = append(results, Result{
				ClientID:  clientID,
				MessageID: messageID,
				Latency:   time.Since(startTime),
				Error:     fmt.Errorf("set deadline: %w", err),
			})

			return results
		}

		expected := fmt.Sprintf(
			"client-%d-message-%d",
			clientID,
			messageID,
		)

		message := []byte(expected)

		if _, err := conn.Write(message); err != nil {
			results = append(results, Result{
				ClientID:  clientID,
				MessageID: messageID,
				Latency:   time.Since(startTime),
				Error:     fmt.Errorf("write: %w", err),
			})

			return results
		}

		response := make([]byte, len(message))

		if _, err := io.ReadFull(conn, response); err != nil {
			results = append(results, Result{
				ClientID:  clientID,
				MessageID: messageID,
				Latency:   time.Since(startTime),
				Error:     fmt.Errorf("read: %w", err),
			})

			return results
		}

		latency := time.Since(startTime)

		if string(response) != expected {
			results = append(results, Result{
				ClientID:  clientID,
				MessageID: messageID,
				Latency:   latency,
				Error: fmt.Errorf(
					"client=%d message=%d expected=%q received=%q",
					clientID,
					messageID,
					expected,
					string(response),
				),
			})

			return results
		}

		results = append(results, Result{
			ClientID:  clientID,
			MessageID: messageID,
			Latency:   latency,
		})
	}

	return results
}