package metrics

import (
	"math/rand/v2"
	"runtime"
)

type Metrics struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

const (
	Alloc         = "Alloc"
	BuckHashSys   = "BuckHashSys"
	Frees         = "Frees"
	GCCPUFraction = "GCCPUFraction"
	GCSys         = "GCSys"
	HeapAlloc     = "HeapAlloc"
	HeapIdle      = "HeapIdle"
	HeapInuse     = "HeapInuse"
	HeapObjects   = "HeapObjects"
	HeapReleased  = "HeapReleased"
	HeapSys       = "HeapSys"
	LastGC        = "LastGC"
	Lookups       = "Lookups"
	MCacheInuse   = "MCacheInuse"
	MSpanSys      = "MSpanSys"
	Mallocs       = "Mallocs"
	NextGC        = "NextGC"
	NumForcedGC   = "NumForcedGC"
	NumGC         = "NumGC"
	OtherSys      = "OtherSys"
	PauseTotalNs  = "PauseTotalNs"
	StackInuse    = "StackInuse"
	StackSys      = "StackSys"
	Sys           = "Sys"
	TotalAlloc    = "TotalAlloc"
	RandomValue   = "RandomValue"
	PollCount     = "PollCount"
)

// var allowedGaugeMetrics = map[string]struct{}{
// 	Alloc:         {},
// 	BuckHashSys:   {},
// 	Frees:         {},
// 	GCCPUFraction: {},
// 	GCSys:         {},
// 	HeapAlloc:     {},
// 	HeapIdle:      {},
// 	HeapInuse:     {},
// 	HeapObjects:   {},
// 	HeapReleased:  {},
// 	HeapSys:       {},
// 	LastGC:        {},
// 	Lookups:       {},
// 	MCacheInuse:   {},
// 	MSpanSys:      {},
// 	Mallocs:       {},
// 	NextGC:        {},
// 	NumForcedGC:   {},
// 	NumGC:         {},
// 	OtherSys:      {},
// 	PauseTotalNs:  {},
// 	StackInuse:    {},
// 	StackSys:      {},
// 	Sys:           {},
// 	TotalAlloc:    {},
// }

// var allowedCounterMetrics = map[string]struct{}{
// 	RandomValue: {},
// 	PollCount:   {},
// }

// func IsMetricGaugeAllowed(metricsName string) bool {
// 	_, exists := allowedGaugeMetrics[metricsName]
// 	return exists
// }

// func IsMetricCounterAllowed(metricsName string) bool {
// 	_, exists := allowedCounterMetrics[metricsName]
// 	return exists
// }

func NewMetrics() *Metrics {
	return &Metrics{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

func (m *Metrics) CollectMetrics() {

	memstats := runtime.MemStats{}
	runtime.ReadMemStats(&memstats)

	m.Gauge[Alloc] = float64(memstats.Alloc)
	m.Gauge[BuckHashSys] = float64(memstats.BuckHashSys)
	m.Gauge[Frees] = float64(memstats.Frees)
	m.Gauge[GCCPUFraction] = memstats.GCCPUFraction
	m.Gauge[GCSys] = float64(memstats.GCSys)
	m.Gauge[HeapAlloc] = float64(memstats.HeapAlloc)
	m.Gauge[HeapIdle] = float64(memstats.HeapIdle)
	m.Gauge[HeapInuse] = float64(memstats.HeapInuse)
	m.Gauge[HeapObjects] = float64(memstats.HeapObjects)
	m.Gauge[HeapReleased] = float64(memstats.HeapReleased)
	m.Gauge[HeapSys] = float64(memstats.HeapSys)
	m.Gauge[LastGC] = float64(memstats.LastGC)
	m.Gauge[Lookups] = float64(memstats.Lookups)
	m.Gauge[MCacheInuse] = float64(memstats.MCacheInuse)
	m.Gauge[MSpanSys] = float64(memstats.MSpanSys)
	m.Gauge[Mallocs] = float64(memstats.Mallocs)
	m.Gauge[NextGC] = float64(memstats.NextGC)
	m.Gauge[NumForcedGC] = float64(memstats.NumForcedGC)
	m.Gauge[NumGC] = float64(memstats.NumGC)
	m.Gauge[OtherSys] = float64(memstats.OtherSys)
	m.Gauge[PauseTotalNs] = float64(memstats.PauseTotalNs)
	m.Gauge[StackInuse] = float64(memstats.StackInuse)
	m.Gauge[StackSys] = float64(memstats.StackSys)
	m.Gauge[Sys] = float64(memstats.Sys)
	m.Gauge[TotalAlloc] = float64(memstats.TotalAlloc)
	m.Gauge[RandomValue] = rand.Float64()
	m.Counter[PollCount] += 1
}
