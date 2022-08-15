package main

import (
	"runtime"

	"github.com/horseinthesky/metricsagent/internal/agent"
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
	agent := agent.New(pollInterval, reportInterval, "")

	for {
		select {
		case <-agent.ReportTicker.C:
			agent.SendMetrics()
		case <-agent.PollTicker.C:
			agent.Count++

			runtime.ReadMemStats(data)

			agent.UpdateMetrics(data)
		}
	}
}
