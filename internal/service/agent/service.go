package service

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/fatkulllin/metrilo/internal/metrics"
)

type MetricsService struct {
	metrics *metrics.Metrics
}

func NewMetricsService(metrics *metrics.Metrics) *MetricsService {
	return &MetricsService{metrics: metrics}
}

func (s *MetricsService) CollectMetrics() {
	s.metrics.CollectMetrics()
}

func (s *MetricsService) GetMetrics() *metrics.Metrics {
	return s.metrics
}

func (s *MetricsService) SendToServer(client *http.Client, method string, endpoint string, reqBody []byte) ([]byte, int) {

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatalf("Error Occurred. %+v", err)
	}
	req.Header.Add("Content-Type", "application/json")
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
	fmt.Printf("Тело ответа: %s\n", body)
	return body, response.StatusCode
}
