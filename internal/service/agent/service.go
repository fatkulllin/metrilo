package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/fatkulllin/metrilo/internal/metrics"
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

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Log.Error("Failed to create request", zap.Error(err))
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Content-Encoding", "gzip")
	req.Header.Add("Content-Type", "application/json")

	secretkey := []byte("secretkey")
	// подписываем алгоритмом HMAC, используя SHA-256
	h := hmac.New(sha256.New, secretkey)
	h.Write([]byte(reqBody))
	sign := h.Sum(nil)

	encodeSign := hex.EncodeToString(sign)

	req.Header.Add("HashSHA256", encodeSign)

	response, err := client.Do(req)
	if err != nil {
		logger.Log.Error("Error sending request", zap.Error(err))
		return err
	}

	// Close the connection to reuse it
	defer response.Body.Close()

	_, err = io.ReadAll(response.Body)
	if err != nil {
		logger.Log.Error("Couldn't parse response body:", zap.String("error", err.Error()))
		return err
	}

	if response.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("server returned status: %d", response.StatusCode)
		logger.Log.Error(errMsg, zap.Int("status", response.StatusCode))
		return errors.New(errMsg)
	}
	logger.Log.Info("Metrics sent successfully")
	return nil
}
