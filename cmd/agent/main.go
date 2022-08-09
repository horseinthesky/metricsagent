package main

import (
	"runtime"
)

// Seconds
const (
	pollInterval   = 2
	reportInterval = 10
)

var (
	data = &runtime.MemStats{}
)

func main() {
	agent := newAgent(pollInterval, reportInterval, "")

	for {
		select {
		case <-agent.report.C:
			agent.sendMetrics()
		case <-agent.poll.C:
			agent.count++

			runtime.ReadMemStats(data)

			agent.updateMetrics()
		}
	}
}
