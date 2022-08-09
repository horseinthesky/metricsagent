package main

import (
	"runtime"
	"time"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
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
