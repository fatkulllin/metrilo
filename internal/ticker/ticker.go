package ticker

import (
	"log"
	"time"

	config "github.com/fatkulllin/metrilo/internal/config/server"
	"github.com/fatkulllin/metrilo/internal/logger"
	service "github.com/fatkulllin/metrilo/internal/service/server"
	"go.uber.org/zap"
)

type Ticker struct {
	StoreInterval   int
	FileStoragePath string
	Restore         bool
	service         *service.MetricsService
	done            chan struct{}
}

func NewTicker(cfg *config.Config, service *service.MetricsService) *Ticker {
	ticker := &Ticker{
		StoreInterval:   cfg.StoreInterval,
		FileStoragePath: cfg.FileStoragePath,
		Restore:         cfg.Restore,
		service:         service,
		done:            make(chan struct{}),
	}
	logger.Log.Info("Store Interval:", zap.Int("storeInterval", ticker.StoreInterval))
	logger.Log.Info("File storage path:", zap.String("server", ticker.FileStoragePath))
	logger.Log.Info("Restore:", zap.Bool("server", ticker.Restore))
	return ticker
}

func (t *Ticker) Start() {
	storeInterval := time.NewTicker(time.Duration(t.StoreInterval) * time.Second)
	defer storeInterval.Stop()
	for {
		select {
		case <-storeInterval.C:
			logger.Log.Info("Save metrics to file")
			err := t.service.SaveMetricsToFile(t.FileStoragePath)
			if err != nil {
				log.Println("Error save metrics", err)
			}
		case <-t.done:
			logger.Log.Info("Ticker stop")
			return
		}
	}
}

func (t *Ticker) Stop() {
	close(t.done)
}
