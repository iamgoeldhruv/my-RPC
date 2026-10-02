package main

import (
	"flag"
	"fmt"
	"time"
)
const (
	serverAddress = "localhost:8080"
	messageCount  = 100
	clientTimeout = 5 * time.Second
	clientCount   = 1000
)

func main() {

	scenarioFlag := flag.String(
		"scenario",
		"normal",
		"test scenario",
	)

	flag.Parse()

	scenario := Scenario(*scenarioFlag)

	fmt.Printf("Running scenario: %s\n", scenario)

	startTime := time.Now()

	result, err := runScenario(scenario)

	if err != nil {
		fmt.Printf("Scenario failed: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("Scenario completed")
	fmt.Printf("Scenario: %s\n", result.Scenario)
	fmt.Printf("Duration: %s\n", result.Duration)

	if result.ServerShutdown != nil {
		fmt.Printf(
			"Server shutdown error: %v\n",
			result.ServerShutdown,
		)
	}

	if result.ClientResults != nil {
		metrics := calculateMetrics(
			result.ClientResults,
			result.Duration,
		)

		printMetrics(metrics)
	}

	_ = startTime
}