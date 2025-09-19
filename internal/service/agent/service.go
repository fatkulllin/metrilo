package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"

	"github.com/fatkulllin/metrilo/internal/metrics"
)

type MetricsService struct {
	metrics *metrics.Metrics
}

func NewMetricsService(metrics *metrics.Metrics) *MetricsService {
	return &MetricsService{metrics: metrics}
}

func (s *MetricsService) CollectMetrics(m *sync.RWMutex) {
	s.metrics.CollectMetrics(m)
}

func (s *MetricsService) CollectGopsutilMetrics(m *sync.RWMutex) {
	s.metrics.CollectGopsutilMetrics(m)
}

func (s *MetricsService) GetMetrics() *metrics.Metrics {
	return s.metrics
}

func (s *MetricsService) SendToServerWithContext(
	ctx context.Context,
	client *http.Client,
	method string,
	endpoint string,
	reqBody []byte,
	wasKeySet bool,
	key []byte,
) error {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Content-Encoding", "gzip")
	req.Header.Add("Content-Type", "application/json")

	// подпись HMAC
	secretkey := []byte("secretkey")
	h := hmac.New(sha256.New, secretkey)
	h.Write(reqBody)
	sign := hex.EncodeToString(h.Sum(nil))
	req.Header.Add("HashSHA256", sign)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}
	return nil
}
