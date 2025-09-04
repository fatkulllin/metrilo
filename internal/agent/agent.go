package agent

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	config "github.com/fatkulllin/metrilo/internal/config/agent"
	"github.com/fatkulllin/metrilo/internal/encoder"
	"github.com/fatkulllin/metrilo/internal/gzip"
	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/fatkulllin/metrilo/internal/models"
	service "github.com/fatkulllin/metrilo/internal/service/agent"
)

type Agent struct {
	ServerAddress  string
	ReportInterval int
	PollInterval   int
	Service        *service.MetricsService
	config         *config.Config
	RateLimit      int
	PublicKey      *rsa.PublicKey
}

func NewAgent(svc *service.MetricsService, cfg *config.Config, publicKey *rsa.PublicKey) *Agent {
	logger.Log.Info("Initializing Agent...")
	agent := &Agent{
		ServerAddress:  cfg.ServerAddress,
		ReportInterval: cfg.ReportInterval,
		PollInterval:   cfg.PollInterval,
		Service:        svc,
		config:         cfg,
		RateLimit:      cfg.RateLimit,
		PublicKey:      publicKey,
	}
	logger.Log.Info("Server address", zap.String("address: ", agent.ServerAddress))
	logger.Log.Info("Report Interval:", zap.Int("report interval: ", agent.ReportInterval))
	logger.Log.Info("Poll Interval:", zap.Int("poll interval: ", agent.PollInterval))
	return agent
}

func newHTTPClient() *http.Client {
	client := &http.Client{}
	return client
}

func (agent *Agent) Run() error {
	pollInterval := time.NewTicker(time.Duration(agent.PollInterval) * time.Second)
	defer pollInterval.Stop()
	reportInterval := time.NewTicker(time.Duration(agent.ReportInterval) * time.Second)
	defer reportInterval.Stop()

	var m sync.RWMutex

	jobs := make(chan []models.Metrics, 5)

	var collected []models.Metrics

	workerCount := agent.config.RateLimit
	for i := range workerCount {
		go agent.worker(i, agent.PublicKey, jobs)
	}

	for {
		select {
		case <-pollInterval.C:
			go agent.Service.CollectMetrics(&m)
			go agent.Service.CollectGopsutilMetrics(&m)
		case <-reportInterval.C:
			m.RLock()
			collected = agent.buildMetrics()
			m.RUnlock()
			batchSize := 10
			for i := 0; i < len(collected); i += batchSize {
				end := min(i+batchSize, len(collected))
				batch := collected[i:end]
				jobs <- batch // Кладём в очередь
			}
			// close(jobs)
		}
	}
}

func (agent *Agent) buildMetrics() []models.Metrics {
	metrics := make([]models.Metrics, 0)
	for k, v := range agent.Service.GetMetrics().Gauge {
		metrics = append(metrics, models.Metrics{
			ID:    k,
			MType: "gauge",
			Value: &v})
	}
	for k, v := range agent.Service.GetMetrics().Counter {
		metrics = append(metrics, models.Metrics{
			ID:    k,
			MType: "counter",
			Delta: &v})
	}
	return metrics
}

func (agent *Agent) worker(id int, publicKey *rsa.PublicKey, jobs <-chan []models.Metrics) {

	endpoint := fmt.Sprintf("http://%v/updates/", agent.ServerAddress)
	client := newHTTPClient()
	label := []byte("agent")
	for batch := range jobs {
		reqBody, err := json.Marshal(batch)
		if err != nil {
			logger.Log.Error("Failed to marshal batch", zap.Error(err))
			continue
		}

		gzipped, err := gzip.GzipCompress(reqBody)
		if err != nil {
			logger.Log.Error("Failed to gzip batch", zap.Error(err))
			continue
		}

		ciphertext, err := encoder.BuildPacket(agent.PublicKey, gzipped, label)
		if err != nil {
			logger.Log.Error("build encode packet error", zap.Error(err))
			continue
		}
		err = agent.Service.SendToServer(client, http.MethodPost, endpoint, ciphertext, agent.config.WasKeySet, []byte(agent.config.Key))
		if err != nil {
			logger.Log.Error("Worker failed to send batch", zap.Int("workerID", id), zap.Error(err))
		} else {
			logger.Log.Info("Worker sent batch", zap.Int("workerID", id), zap.Int("batchSize", len(batch)))
		}
	}
}
