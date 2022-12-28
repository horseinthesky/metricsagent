package agent

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// Agent custom metric types
type (
	gauge   = float64
	counter = int64
)

// updatePSUtilMetrics updates psutil metrics.
func updatePSUtilMetrics(storage *sync.Map) {
	memory, _ := mem.VirtualMemory()
	storage.Store("TotalMemory", gauge(memory.Total))
	storage.Store("FreeMemory", gauge(memory.Free))

	cpusUtilization, _ := cpu.Percent(0, true)
	for i, c := range cpusUtilization {
		storage.Store(fmt.Sprintf("CPUutilization%d", i), gauge(c))
	}
}

// updateRuntimeMetrics updates runtime metrics.
func updateRuntimeMetrics(storage *sync.Map) {
	data := &runtime.MemStats{}

	runtime.ReadMemStats(data)

	storage.Store("Alloc", gauge(data.Alloc))
	storage.Store("BuckHashSys", gauge(data.BuckHashSys))
	storage.Store("Frees", gauge(data.Frees))
	storage.Store("GCCPUFraction", gauge(data.GCCPUFraction))
	storage.Store("GCSys", gauge(data.GCSys))
	storage.Store("HeapAlloc", gauge(data.HeapAlloc))
	storage.Store("HeapIdle", gauge(data.HeapIdle))
	storage.Store("HeapInuse", gauge(data.HeapInuse))
	storage.Store("HeapObjects", gauge(data.HeapObjects))
	storage.Store("HeapReleased", gauge(data.HeapReleased))
	storage.Store("HeapSys", gauge(data.HeapSys))
	storage.Store("LastGC", gauge(data.LastGC))
	storage.Store("Lookups", gauge(data.Lookups))
	storage.Store("MCacheInuse", gauge(data.MCacheInuse))
	storage.Store("MCacheSys", gauge(data.MCacheSys))
	storage.Store("MSpanInuse", gauge(data.MSpanInuse))
	storage.Store("MSpanSys", gauge(data.MSpanSys))
	storage.Store("Mallocs", gauge(data.Mallocs))
	storage.Store("NextGC", gauge(data.NextGC))
	storage.Store("NumForcedGC", gauge(data.NumForcedGC))
	storage.Store("NumGC", gauge(data.NumGC))
	storage.Store("OtherSys", gauge(data.OtherSys))
	storage.Store("PauseTotalNs", gauge(data.PauseTotalNs))
	storage.Store("StackInuse", gauge(data.StackInuse))
	storage.Store("StackSys", gauge(data.StackSys))
	storage.Store("Sys", gauge(data.Sys))
	storage.Store("TotalAlloc", gauge(data.TotalAlloc))
	storage.Store("RandomValue", gauge(rand.Float64()))
}
