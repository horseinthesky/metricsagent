package agent

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// Agent custom metric types
type (
	gauge   = float64
	counter = int64
)

// collectPSUtilMetrics runs updatePSUtilMetrics every config.PollInterval.
// Also handles graceful shutdown.
func (a *Agent) collectPSUtilMetrics(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("psutil data collection cancelled")
			return
		case <-a.PollTicker.C:
			a.updatePSUtilMetrics()

			log.Println("successfully collected psutil metrics")
		}
	}
}

// collectRuntimeMetrics runs updateRuntimeMetrics every config.PollInterval.
// Also handles graceful shutdown.
func (a *Agent) collectRuntimeMetrics(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("runtime data collection cancelled")
			return
		case <-a.PollTicker.C:
			a.updateRuntimeMetrics()

			log.Println("successfully collected runtime metrics")
		}
	}
}

// updatePSUtilMetrics updates psutil metrics.
func (a *Agent) updatePSUtilMetrics() {
	memory, _ := mem.VirtualMemory()
	a.metrics.Store("TotalMemory", gauge(memory.Total))
	a.metrics.Store("FreeMemory", gauge(memory.Free))

	cpusUtilization, _ := cpu.Percent(0, true)
	for i, c := range cpusUtilization {
		a.metrics.Store(fmt.Sprintf("CPUutilization%d", i), gauge(c))
	}
}

// updateRuntimeMetrics updates runtime metrics.
func (a *Agent) updateRuntimeMetrics() {
	data := &runtime.MemStats{}

	a.PollCounter++

	runtime.ReadMemStats(data)

	a.metrics.Store("Alloc", gauge(data.Alloc))
	a.metrics.Store("BuckHashSys", gauge(data.BuckHashSys))
	a.metrics.Store("Frees", gauge(data.Frees))
	a.metrics.Store("GCCPUFraction", gauge(data.GCCPUFraction))
	a.metrics.Store("GCSys", gauge(data.GCSys))
	a.metrics.Store("HeapAlloc", gauge(data.HeapAlloc))
	a.metrics.Store("HeapIdle", gauge(data.HeapIdle))
	a.metrics.Store("HeapInuse", gauge(data.HeapInuse))
	a.metrics.Store("HeapObjects", gauge(data.HeapObjects))
	a.metrics.Store("HeapReleased", gauge(data.HeapReleased))
	a.metrics.Store("HeapSys", gauge(data.HeapSys))
	a.metrics.Store("LastGC", gauge(data.LastGC))
	a.metrics.Store("Lookups", gauge(data.Lookups))
	a.metrics.Store("MCacheInuse", gauge(data.MCacheInuse))
	a.metrics.Store("MCacheSys", gauge(data.MCacheSys))
	a.metrics.Store("MSpanInuse", gauge(data.MSpanInuse))
	a.metrics.Store("MSpanSys", gauge(data.MSpanSys))
	a.metrics.Store("Mallocs", gauge(data.Mallocs))
	a.metrics.Store("NextGC", gauge(data.NextGC))
	a.metrics.Store("NumForcedGC", gauge(data.NumForcedGC))
	a.metrics.Store("NumGC", gauge(data.NumGC))
	a.metrics.Store("OtherSys", gauge(data.OtherSys))
	a.metrics.Store("PauseTotalNs", gauge(data.PauseTotalNs))
	a.metrics.Store("StackInuse", gauge(data.StackInuse))
	a.metrics.Store("StackSys", gauge(data.StackSys))
	a.metrics.Store("Sys", gauge(data.Sys))
	a.metrics.Store("TotalAlloc", gauge(data.TotalAlloc))
	a.metrics.Store("RandomValue", gauge(rand.Float64()))
}
