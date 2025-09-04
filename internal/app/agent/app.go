package app

import (
	"github.com/fatkulllin/metrilo/internal/agent"
	config "github.com/fatkulllin/metrilo/internal/config/agent"
	"github.com/fatkulllin/metrilo/internal/keysmanager"
	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/fatkulllin/metrilo/internal/metrics"
	service "github.com/fatkulllin/metrilo/internal/service/agent"
	"go.uber.org/zap"
)

type App struct {
	metrics *metrics.Metrics
	agent   *agent.Agent
	service *service.MetricsService
}

func NewApp(cfg *config.Config) *App {
	publicKey, err := keysmanager.LoadPublicKey(cfg.CryptoKey)
	if err != nil {
		logger.Log.Fatal("не удалось получить публичный ключ", zap.Error(err))
	}
	metrics := metrics.NewMetrics()
	service := service.NewMetricsService(metrics)
	agent := agent.NewAgent(service, cfg, publicKey)

	return &App{
		metrics: metrics,
		service: service,
		agent:   agent,
	}
}

func (a *App) Run() {
	a.agent.Run()
}
