package app

import (
	config "github.com/fatkulllin/metrilo/internal/config/server"
	"github.com/fatkulllin/metrilo/internal/handlers"
	"github.com/fatkulllin/metrilo/internal/server"
	"github.com/fatkulllin/metrilo/internal/service"
	"github.com/fatkulllin/metrilo/internal/storage"
)

type App struct {
	memStore *storage.MemStorage
	service  *service.MetricsService
	handlers *handlers.Handlers
	server   *server.Server
}

func NewApp(cfg *config.Config) *App {
	memStore := storage.NewMemoryStorage()
	service := service.NewMetricsService(memStore)
	handlers := handlers.NewHandlers(service)
	server := server.NewServer(handlers, cfg)
	return &App{
		memStore: memStore,
		service:  service,
		handlers: handlers,
		server:   server}
}

func (a *App) Run() {
	a.server.Start()
}
