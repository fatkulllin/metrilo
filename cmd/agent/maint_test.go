package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fatkulllin/metrilo/internal/metrics"
	"github.com/stretchr/testify/assert"
)

var allowedGaugeMetrics = map[string]struct{}{
	metrics.Alloc:         {},
	metrics.BuckHashSys:   {},
	metrics.Frees:         {},
	metrics.GCCPUFraction: {},
	metrics.GCSys:         {},
	metrics.HeapAlloc:     {},
	metrics.HeapIdle:      {},
	metrics.HeapInuse:     {},
	metrics.HeapObjects:   {},
	metrics.HeapReleased:  {},
	metrics.HeapSys:       {},
	metrics.LastGC:        {},
	metrics.Lookups:       {},
	metrics.MCacheInuse:   {},
	metrics.MSpanSys:      {},
	metrics.Mallocs:       {},
	metrics.NextGC:        {},
	metrics.NumForcedGC:   {},
	metrics.NumGC:         {},
	metrics.OtherSys:      {},
	metrics.PauseTotalNs:  {},
	metrics.StackInuse:    {},
	metrics.StackSys:      {},
	metrics.Sys:           {},
	metrics.TotalAlloc:    {},
	metrics.RandomValue:   {},
}

var allowedCounterMetrics = map[string]struct{}{
	metrics.PollCount: {},
}

func TestColleMetrics(t *testing.T) {
	t.Run("Check counter", func(t *testing.T) {
		metriki := metrics.NewMetrics()
		metriki.CollectMetrics()
		for k := range metriki.Counter {
			_, exists := allowedCounterMetrics[k]
			assert.True(t, exists, "Ключ %v должен существовать", k)
		}
		for k := range metriki.Gauge {
			_, exists := allowedGaugeMetrics[k]
			assert.True(t, exists, "Ключ %+v должен существовать", k)
		}

	})
}

func TestSendRequest_ServerError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/plain")
		if req.Method != http.MethodPost {
			http.Error(res, "Only POST requests are allowed!!", http.StatusMethodNotAllowed)
			return
		}
		if req.Header.Get("Content-Type") != "text/plain" {
			http.Error(res, "Only Content-Type: text/plain header are allowed!!", http.StatusMethodNotAllowed)
			return
		}
		res.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	client := HttpClient()
	SendRequest(client, http.MethodPost, "localhost:8080")

}
