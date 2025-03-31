package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

func httpClient() *http.Client {
	client := &http.Client{Timeout: 10 * time.Second}
	return client
}

func main() {
	metrics := &Metrcics{}
	metrics.Gauge = make(map[string]float64, 27)
	metrics.Counter = make(map[string]int64)
	pollInterval := time.NewTicker(time.Duration(2) * time.Second)
	reportInterval := time.NewTicker(time.Duration(3) * time.Second)
	// lastSendMetricsTime := time.Now().Second()
	endpoint := ""
	c := httpClient()

	defer pollInterval.Stop()
	defer reportInterval.Stop()

	for {
		select {
		case <-pollInterval.C:
			metrics.collectMetrics()
		case <-reportInterval.C:
			jsonMetrics, err := json.Marshal(*metrics)
			if err != nil {
				log.Fatalf("Error parse json metrics. %+v", err)
			}
			for k, v := range metrics.Gauge {
				endpoint = fmt.Sprintf("http://localhost:8080/update/gauge/%v/%v", k, v)
				sendRequest(c, http.MethodPost, endpoint, jsonMetrics)
			}
			for k, v := range metrics.Counter {
				endpoint = fmt.Sprintf("http://localhost:8080/update/counter/%v/%v", k, v)
				sendRequest(c, http.MethodPost, endpoint, jsonMetrics)
			}
		}

	}
}

func sendRequest(client *http.Client, method string, endpoint string, jsonData []byte) []byte {

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error Occurred. %+v", err)
	}
	req.Header.Add("Content-Type", "text/plain")

	response, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request to API endpoint. %+v", err)
	}

	// Close the connection to reuse it
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Couldn't parse response body. %+v", err)
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Request failed with status: %s", response.Status)
	}

	return body
}

type Metrcics struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

func (metrics *Metrcics) collectMetrics() {

	memstats := runtime.MemStats{}
	runtime.ReadMemStats(&memstats)
	metrics.Gauge["Alloc"] = float64(memstats.Alloc)
	metrics.Gauge["BuckHashSys"] = float64(memstats.BuckHashSys)
	metrics.Gauge["Frees"] = float64(memstats.Frees)
	metrics.Gauge["GCCPUFraction"] = memstats.GCCPUFraction
	metrics.Gauge["GCSys"] = float64(memstats.GCSys)
	metrics.Gauge["HeapAlloc"] = float64(memstats.HeapAlloc)
	metrics.Gauge["HeapIdle"] = float64(memstats.HeapIdle)
	metrics.Gauge["HeapInuse"] = float64(memstats.HeapInuse)
	metrics.Gauge["HeapObjects"] = float64(memstats.HeapObjects)
	metrics.Gauge["HeapReleased"] = float64(memstats.HeapReleased)
	metrics.Gauge["HeapSys"] = float64(memstats.HeapSys)
	metrics.Gauge["LastGC"] = float64(memstats.LastGC)
	metrics.Gauge["Lookups"] = float64(memstats.Lookups)
	metrics.Gauge["MCacheInuse"] = float64(memstats.MCacheInuse)
	metrics.Gauge["MSpanSys"] = float64(memstats.MSpanSys)
	metrics.Gauge["Mallocs"] = float64(memstats.Mallocs)
	metrics.Gauge["NextGC"] = float64(memstats.NextGC)
	metrics.Gauge["NumForcedGC"] = float64(memstats.NumForcedGC)
	metrics.Gauge["NumGC"] = float64(memstats.NumGC)
	metrics.Gauge["OtherSys"] = float64(memstats.OtherSys)
	metrics.Gauge["PauseTotalNs"] = float64(memstats.PauseTotalNs)
	metrics.Gauge["StackInuse"] = float64(memstats.StackInuse)
	metrics.Gauge["StackSys"] = float64(memstats.StackSys)
	metrics.Gauge["Sys"] = float64(memstats.Sys)
	metrics.Gauge["TotalAlloc"] = float64(memstats.TotalAlloc)
	metrics.Gauge["RandomValue"] = rand.Float64()
	metrics.Counter["PollCount"] += 1
}
