package main

import (
	"sort"
	"time"
)

type Metrics struct {
	Connections    int
	Messages       int
	Successful     int
	Failed         int
	Duration       time.Duration
	Throughput     float64
	AverageLatency time.Duration
	P50            time.Duration
	P95            time.Duration
	P99            time.Duration
	MaxLatency     time.Duration
}

func calculateMetrics(
	results []Result,
	duration time.Duration,
) Metrics {
	metrics := Metrics{
		Connections: clientCount,
		Duration:    duration,
	}

	latencies := make([]time.Duration, 0, len(results))

	for _, result := range results {
		metrics.Messages++

		if result.Error != nil {
			metrics.Failed++
			continue
		}

		metrics.Successful++
		latencies = append(latencies, result.Latency)
	}

	if duration > 0 {
		metrics.Throughput =
			float64(metrics.Messages) / duration.Seconds()
	}

	if len(latencies) == 0 {
		return metrics
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	var totalLatency time.Duration

	for _, latency := range latencies {
		totalLatency += latency
	}

	metrics.AverageLatency =
		totalLatency / time.Duration(len(latencies))

	metrics.P50 = percentile(latencies, 0.50)
	metrics.P95 = percentile(latencies, 0.95)
	metrics.P99 = percentile(latencies, 0.99)
	metrics.MaxLatency = latencies[len(latencies)-1]

	return metrics
}

func percentile(
	latencies []time.Duration,
	percentile float64,
) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	index := int(
		float64(len(latencies)-1) * percentile,
	)

	return latencies[index]
}