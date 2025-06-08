package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	config "github.com/fatkulllin/metrilo/internal/config/agent"
	"github.com/fatkulllin/metrilo/internal/gzip"
	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/fatkulllin/metrilo/internal/models"
	service "github.com/fatkulllin/metrilo/internal/service/agent"
	"go.uber.org/zap"
)

type Agent struct {
	ServerAddress  string
	ReportInterval int
	PollInterval   int
	Service        *service.MetricsService
}

func NewAgent(svc *service.MetricsService, cfg *config.Config) *Agent {
	logger.Log.Info("Initializing Agent...")
	agent := &Agent{
		ServerAddress:  cfg.ServerAddress,
		ReportInterval: cfg.ReportInterval,
		PollInterval:   cfg.PollInterval,
		Service:        svc,
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
	endpoint := fmt.Sprintf("http://%v/updates/", agent.ServerAddress)
	client := newHTTPClient()

	for {
		select {
		case <-pollInterval.C:
			agent.Service.CollectMetrics()
		case <-reportInterval.C:
			metrics := make([]models.Metrics, 0)
			for k, v := range agent.Service.GetMetrics().Gauge {
				metrics = append(metrics, models.Metrics{
					ID:    k,
					MType: "gauge",
					Value: &v})
			}
			for k, v := range agent.Service.GetMetrics().Counter {
				fmt.Println(k, v)
				metrics = append(metrics, models.Metrics{
					ID:    k,
					MType: "counter",
					Delta: &v})
			}
			reqBody, err := json.Marshal(metrics)
			if err != nil {
				logger.Log.Error(err.Error())
			}
			bodyBuf, err := gzip.GzipCompress(reqBody)
			if err != nil {
				logger.Log.Error("Error compress gague body", zap.String("error", err.Error()), zap.String("request body", string(reqBody)))
				return nil
			}
			err = agent.Service.SendToServer(client, http.MethodPost, endpoint, bodyBuf)
			if err != nil {
				logger.Log.Error("Failed to send metrics after retries",
					zap.Error(err),
					zap.String("endpoint", endpoint))
			}
		}
	}
}
