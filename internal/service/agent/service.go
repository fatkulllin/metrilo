package service

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/fatkulllin/metrilo/internal/metrics"
	"github.com/fatkulllin/metrilo/internal/retry"
	"go.uber.org/zap"
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

func (s *MetricsService) SendToServer(client *http.Client, method string, endpoint string, reqBody []byte) error {

	return retry.Do(3, func() error {
		req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(reqBody))
		if err != nil {
			log.Fatalf("Error Occurred. %+v", err)
		}
		req.Header.Add("Content-Encoding", "gzip")
		req.Header.Add("Content-Type", "application/json")
		response, err := client.Do(req)
		if err != nil {
			log.Printf("Error sending request to API endpoint. %+v", err)
			// return nil, 0
			return err
		}

		// Close the connection to reuse it
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			logger.Log.Error("Couldn't parse response body:", zap.String("error", err.Error()))
			return err
		}

		if response.StatusCode != http.StatusOK {
			err := fmt.Errorf("request failed with: %d", response.StatusCode)
			logger.Log.Error("request failed with", zap.Int("status", response.StatusCode))
			return err
		}
		fmt.Printf("Тело ответа: %s\n%d", body, response.StatusCode)
		// return body, response.StatusCode
		return nil
	}, retry.IsNetworkError)
}
