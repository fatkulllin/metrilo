package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	config "github.com/fatkulllin/metrilo/internal/config/server"
	"github.com/fatkulllin/metrilo/internal/database"
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
	db       *database.Database
}

func NewApp(cfg *config.Config) *App {
	memStore := storage.NewMemoryStorage()
	var db *database.Database
	var err error
	if cfg.WasDatabaseSet {
		db, err = database.NewDatabase(cfg.Database)
		if err != nil {
			logger.Log.Warn("Error connect to DB", zap.String("error", err.Error()))
			db = nil
		}
	}

	service := service.NewMetricsService(memStore, cfg, db)
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

	if db != nil {
		if migrateConnect, _ := db.GetDB(); migrateConnect != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			// не забываем освободить ресурс
			defer cancel()
			_, err := migrateConnect.QueryContext(ctx, "CREATE TABLE IF NOT EXISTS counter(name varchar(40) primary key, value integer);")
			if err != nil {
				logger.Log.Error(err.Error())
			}
			_, err = migrateConnect.QueryContext(ctx, "CREATE TABLE IF NOT EXISTS gauge(name varchar(40) primary key, value double precision);")
			if err != nil {
				logger.Log.Error(err.Error())
			}
		}
	}

	return &App{
		memStore: memStore,
		service:  service,
		handlers: handlers,
		server:   server,
		ticker:   tick,
		db:       db,
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

	if a.db != nil {
		if err := a.db.Close(); err != nil {
			logger.Log.Error("Error closing DB", zap.String("error", err.Error()))
		}
		logger.Log.Info("Successfully closed DB connection")
	}

	logger.Log.Info("Graceful shutdown")
}
