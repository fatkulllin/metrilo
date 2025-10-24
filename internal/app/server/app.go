package app

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	config "github.com/fatkulllin/metrilo/internal/config/server"
	"github.com/fatkulllin/metrilo/internal/database"
	gprchandlers "github.com/fatkulllin/metrilo/internal/handlers/gprc"
	httphandlers "github.com/fatkulllin/metrilo/internal/handlers/http"
	"github.com/fatkulllin/metrilo/internal/keysmanager"
	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/fatkulllin/metrilo/internal/retry"
	"github.com/fatkulllin/metrilo/internal/server"
	service "github.com/fatkulllin/metrilo/internal/service/server"
	"github.com/fatkulllin/metrilo/internal/storage"
	"github.com/fatkulllin/metrilo/internal/ticker"
)

type App struct {
	memStore *storage.MemStorage
	service  *service.MetricsService
	handlers *httphandlers.Handlers
	server   *server.Server
	ticker   *ticker.Ticker
	db       *database.Database
}

func NewApp(cfg *config.Config) *App {
	privateKey, err := keysmanager.LoadPrivateKey(cfg.CryptoKey)
	if err != nil {
		logger.Log.Fatal("не удалось получить приватный ключ", zap.Error(err))
	}

	memStore := storage.NewMemoryStorage()
	var db *database.Database
	if cfg.WasDatabaseSet {
		retry.Do(3, func() error {
			db, err = database.NewDatabase(cfg.Database)
			if err != nil {
				logger.Log.Warn("Error connect to DB", zap.String("error", err.Error()))
				db = nil
				return err
			}
			return nil
		}, retry.IsPGError)
	}

	service := service.NewMetricsService(memStore, cfg, db)
	handlers := httphandlers.NewHandlers(service)
	grpcHandlers := gprchandlers.NewMetricsGRPCServer(service)
	server := server.NewServer(handlers, grpcHandlers, cfg, privateKey)

	var tick *ticker.Ticker

	if cfg.StoreInterval > 0 {
		tick = ticker.NewTicker(cfg, service)
	}

	if cfg.Restore {
		err := service.ReadMetricsFromFile(cfg.FileStoragePath)
		if err != nil {
			log.Println("error read metrics from file", err)
		}
		log.Println("read metrics from file okay")
	}

	if db != nil {
		if migrateConnect, _ := db.GetDB(); migrateConnect != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			// не забываем освободить ресурс
			defer cancel()
			_, err := migrateConnect.QueryContext(ctx, "CREATE TABLE IF NOT EXISTS counter(name varchar(40) primary key, value bigint);")
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

func (a *App) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 3)
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := a.server.Start(ctx); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("server exited with error", zap.Error(err))
			errCh <- err
		}
	}()

	if a.ticker != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.ticker.Start(ctx)
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.server.StartGRPC(ctx); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("server exited with error", zap.Error(err))
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Log.Info("context canceled, shutting down...")
		// сохранить метрики
		if err := a.service.SaveMetricsToFile(".temp"); err != nil {
			logger.Log.Error("Error save metrics to file", zap.String("error", err.Error()))
		} else {
			logger.Log.Info("Successfully saved metrics to file")
		}

		// закрыть БД
		if a.db != nil {
			if err := a.db.Close(); err != nil {
				logger.Log.Error("Error closing DB", zap.String("error", err.Error()))
			} else {
				logger.Log.Info("Successfully closed DB connection")
			}
		}
	case err := <-errCh:
		logger.Log.Warn("shutting down due to error")
		return err
	}

	wg.Wait()
	logger.Log.Info("shutdown complete")
	return nil
}
