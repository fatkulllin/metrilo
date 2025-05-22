package app

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	config "github.com/fatkulllin/metrilo/internal/config/server"
	"github.com/fatkulllin/metrilo/internal/handlers"
	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/fatkulllin/metrilo/internal/server"
	service "github.com/fatkulllin/metrilo/internal/service/server"
	"github.com/fatkulllin/metrilo/internal/storage"
	"github.com/fatkulllin/metrilo/internal/ticker"
	"go.uber.org/zap"
)

type App struct {
	memStore *storage.MemStorage
	service  *service.MetricsService
	handlers *handlers.Handlers
	server   *server.Server
	ticker   *ticker.Ticker
}

func NewApp(cfg *config.Config) *App {
	memStore := storage.NewMemoryStorage()
	service := service.NewMetricsService(memStore, cfg.StoreInterval, cfg.FileStoragePath)
	handlers := handlers.NewHandlers(service)
	server := server.NewServer(handlers, cfg)

	var tick *ticker.Ticker

	if cfg.StoreInterval > 0 {
		tick = ticker.NewTicker(cfg, service)
	}

	if cfg.Restore {
		err := service.ReadMetricsFromFile(cfg.FileStoragePath)
		if err != nil {
			log.Println("error read metrics from file", err)
		}
		log.Println("Read metrics from file okay")
	}

	return &App{
		memStore: memStore,
		service:  service,
		handlers: handlers,
		server:   server,
		ticker:   tick,
	}
}

func (a *App) Run() {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go a.server.Start()

	if a.ticker != nil {
		go a.ticker.Start()
	}

	sig := <-sigs

	logger.Log.Info("Get syscall", zap.String("syscall", sig.String()))

	if a.ticker != nil {
		a.ticker.Stop()
	}

	err := a.service.SaveMetricsToFile(".temp")

	if err != nil {
		logger.Log.Error("Error save metrics to file", zap.String("error", err.Error()))
	}
	logger.Log.Info("Successfully save metrics to file")
	logger.Log.Info("graceful shutdown")
}
