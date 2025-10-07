package ticker

import (
	"context"
	"log"
	"time"

	"go.uber.org/zap"

	config "github.com/fatkulllin/metrilo/internal/config/server"
	"github.com/fatkulllin/metrilo/internal/logger"
	service "github.com/fatkulllin/metrilo/internal/service/server"
)

type Ticker struct {
	StoreInterval   int
	FileStoragePath string
	Restore         bool
	service         *service.MetricsService
}

func NewTicker(cfg *config.Config, service *service.MetricsService) *Ticker {
	ticker := &Ticker{
		StoreInterval:   cfg.StoreInterval,
		FileStoragePath: cfg.FileStoragePath,
		Restore:         cfg.Restore,
		service:         service,
	}
	logger.Log.Info("Store Interval", zap.Int("storeInterval", ticker.StoreInterval))
	logger.Log.Info("File storage path", zap.String("server", ticker.FileStoragePath))
	logger.Log.Info("Restore: ", zap.Bool("server", ticker.Restore))
	return ticker
}

func (t *Ticker) Start(ctx context.Context) {
	storeInterval := time.NewTicker(time.Duration(t.StoreInterval) * time.Second)
	defer storeInterval.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Ticker stop")
			return
		case <-storeInterval.C:
			logger.Log.Info("Save metrics to file")
			err := t.service.SaveMetricsToFile(t.FileStoragePath)
			if err != nil {
				log.Println("Error save metrics", err)
			}
		}
	}
}
