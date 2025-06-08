package app

import (
	"github.com/fatkulllin/metrilo/internal/agent"
	config "github.com/fatkulllin/metrilo/internal/config/agent"
	"github.com/fatkulllin/metrilo/internal/metrics"
	service "github.com/fatkulllin/metrilo/internal/service/agent"
)

type App struct {
	metrics *metrics.Metrics
	agent   *agent.Agent
	service *service.MetricsService
}

func NewApp(cfg *config.Config) *App {
	metrics := metrics.NewMetrics()
	service := service.NewMetricsService(metrics)
	agent := agent.NewAgent(service, cfg)

	return &App{
		metrics: metrics,
		service: service,
		agent:   agent,
	}
}

func (a *App) Run() error {
	err := a.agent.Run()
	return err
}
