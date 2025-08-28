package metrics

import (
	"sync"
	"testing"
)

func BenchmarkCollectMetrics(b *testing.B) {
	var m sync.RWMutex
	metrics := NewMetrics()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		metrics.CollectMetrics(&m)
	}
}

func BenchmarkCollectGopsutilMetrics(b *testing.B) {
	var m sync.RWMutex
	metrics := NewMetrics()

	b.Run("loop", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := metrics.CollectGopsutilMetrics(&m); err != nil {
				b.Fatal(err)
			}
		}
		b.Logf("Gauge size after benchmark: %d", len(metrics.Gauge))
	})
}
