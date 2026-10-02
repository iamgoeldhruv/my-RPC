package main

import "fmt"

func printMetrics(metrics Metrics) {
	fmt.Println()
	fmt.Println("Load Test Results")
	fmt.Println("=================")

	fmt.Printf("Connections:       %d\n", metrics.Connections)
	fmt.Printf("Messages:          %d\n", metrics.Messages)
	fmt.Printf("Successful:        %d\n", metrics.Successful)
	fmt.Printf("Failed:            %d\n", metrics.Failed)

	fmt.Println()

	fmt.Printf("Duration:          %s\n", metrics.Duration)
	fmt.Printf("Throughput:        %.2f msg/s\n", metrics.Throughput)

	fmt.Println()

	fmt.Printf("Average latency:   %s\n", metrics.AverageLatency)
	fmt.Printf("P50:               %s\n", metrics.P50)
	fmt.Printf("P95:               %s\n", metrics.P95)
	fmt.Printf("P99:               %s\n", metrics.P99)
	fmt.Printf("Max:               %s\n", metrics.MaxLatency)
}