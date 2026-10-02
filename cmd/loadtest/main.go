package main

import (
	"sync"
	"time"
)

const (
	serverAddress = "localhost:8080"
	messageCount  = 100
	clientTimeout = 5 * time.Second
	clientCount   = 1000
)

func main() {
	startTime := time.Now()

	var wg sync.WaitGroup

	results := make(chan []Result, clientCount)

	for clientID := 1; clientID <= clientCount; clientID++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			results <- runClient(serverAddress, id)
		}(clientID)
	}

	wg.Wait()
	close(results)

	var allResults []Result

	for clientResults := range results {
		allResults = append(allResults, clientResults...)
	}

	duration := time.Since(startTime)

	metrics := calculateMetrics(allResults, duration)

	printMetrics(metrics)
}