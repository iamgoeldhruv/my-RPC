package main

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"my-rpc/internal/framing"
)

type Result struct {
	ClientID  int
	MessageID int
	Latency   time.Duration
	Error     error
}

type ClientEventType int

const (
	ClientConnected ClientEventType = iota
	MessageCompleted
	ClientFinished
)

type ClientEvent struct {
	ClientID  int
	MessageID int
	Type      ClientEventType
}

func runClient(
	address string,
	clientID int,
	eventCh chan<- ClientEvent,
	startMessages <-chan struct{},
	stopAfter int,
) []Result {
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

	eventCh <- ClientEvent{
		ClientID: clientID,
		Type:     ClientConnected,
	}

	// Used by shutdown scenarios.
	// For the normal load test, this channel is nil.
	if startMessages != nil {
		<-startMessages
	}

	for messageID := 1; messageID <= messageCount; messageID++ {
		// Used to deliberately stop clients after N messages.
		if stopAfter > 0 && messageID > stopAfter {
			return results
		}

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

		request := []byte(expected)

		// Encode and write the complete request frame.
		if err := framing.WriteFrame(conn, request); err != nil {
			results = append(results, Result{
				ClientID:  clientID,
				MessageID: messageID,
				Latency:   time.Since(startTime),
				Error:     fmt.Errorf("write request frame: %w", err),
			})
			return results
		}

		// Read exactly one complete response frame.
		response, err := framing.ReadFrame(conn)
		if err != nil {
			results = append(results, Result{
				ClientID:  clientID,
				MessageID: messageID,
				Latency:   time.Since(startTime),
				Error:     fmt.Errorf("read response frame: %w", err),
			})
			return results
		}

		latency := time.Since(startTime)

		// Compare bytes rather than strings so binary payloads are supported.
		if !bytes.Equal(response, request) {
			results = append(results, Result{
				ClientID:  clientID,
				MessageID: messageID,
				Latency:   latency,
				Error: fmt.Errorf(
					"client=%d message=%d expected=%q received=%q",
					clientID,
					messageID,
					request,
					response,
				),
			})
			return results
		}

		results = append(results, Result{
			ClientID:  clientID,
			MessageID: messageID,
			Latency:   latency,
		})

		eventCh <- ClientEvent{
			ClientID:  clientID,
			MessageID: messageID,
			Type:      MessageCompleted,
		}
	}

	eventCh <- ClientEvent{
		ClientID: clientID,
		Type:     ClientFinished,
	}

	return results
}
