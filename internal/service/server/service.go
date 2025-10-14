package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	config "github.com/fatkulllin/metrilo/internal/config/server"
	"github.com/fatkulllin/metrilo/internal/database"
	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/fatkulllin/metrilo/internal/metrics"
	"github.com/fatkulllin/metrilo/internal/models"
	"github.com/fatkulllin/metrilo/internal/storage"
)

type MetricsService struct {
	store  *storage.MemStorage
	config *config.Config
	db     *database.Database
}

func NewMetricsService(store *storage.MemStorage, config *config.Config, db *database.Database) *MetricsService {
	return &MetricsService{store: store, config: config, db: db}
}

func (s *MetricsService) SaveGauge(name string, value float64, ctx context.Context) error {

	if s.config.WasDatabaseSet {
		dbConnect, err := s.db.GetDB()
		if err != nil {
			logger.Log.Error("Can not get DB connection", zap.Error(err))
			return err
		}
		logger.Log.Info("Save metric to DB", zap.String("gauge", name))
		return s.store.SaveGaugeToDB(dbConnect, name, value, ctx)
	}

	s.store.SaveGauge(name, value)

	if s.config.StoreInterval == 0 || (s.config.WasPathSet && s.config.WasIntervalSet) {
		logger.Log.Info("Saving gague to file")
		if err := s.SaveMetricsToFile(s.config.FileStoragePath); err != nil {
			logger.Log.Error("Failed to save to file", zap.Error(err))
			return err
		}
		return s.SaveMetricsToFile(s.config.FileStoragePath)
	}

	return nil
}

func (s *MetricsService) SaveCounter(name string, delta int64, ctx context.Context) error {

	if s.config.WasDatabaseSet {
		dbConnect, err := s.db.GetDB()
		if err != nil {
			logger.Log.Error("Can not get DB connection", zap.Error(err))
			return err
		}
		logger.Log.Info("Save metric to DB", zap.String("counter", name), zap.Int64("value", delta))
		return s.store.SaveCounterToDB(dbConnect, name, delta, ctx)
	}

	s.store.SaveCounter(name, delta)

	if s.config.StoreInterval == 0 || (s.config.WasPathSet && s.config.WasIntervalSet) {
		logger.Log.Info("Saving gague to file")
		if err := s.SaveMetricsToFile(s.config.FileStoragePath); err != nil {
			logger.Log.Error("Failed to save to file", zap.Error(err))
			return err
		}
		return s.SaveMetricsToFile(s.config.FileStoragePath)
	}

	return nil
}

func (s *MetricsService) GetCounter(name string, ctx context.Context) (int64, error) {
	if s.config.WasDatabaseSet {
		dbConnect, err := s.db.GetDB()
		if err != nil {
			logger.Log.Error("Can not get DB connection", zap.Error(err))
			return 0, err
		}
		logger.Log.Info("Get metric to DB", zap.String("counter", name))
		return s.store.GetCounterDB(dbConnect, name, ctx)
	}
	return s.store.GetCounter(name)
}

func (s *MetricsService) GetGauge(name string, ctx context.Context) (float64, error) {
	if s.config.WasDatabaseSet {
		dbConnect, err := s.db.GetDB()
		if err != nil {
			logger.Log.Error("Can not get DB connection", zap.Error(err))
			return 0, err
		}
		logger.Log.Info("Get metric to DB", zap.String("counter", name))
		return s.store.GetGaugeDB(dbConnect, name, ctx)
	}
	return s.store.GetGauge(name)
}
func (s *MetricsService) GetMetrics() (map[string]float64, map[string]int64) {
	return s.store.GetMetrics()
}

func (s *MetricsService) SaveMetricsToFile(filename string) error {
	gauges, counters := s.store.GetMetrics()
	metrics := storage.MemStorage{
		Gauge:   gauges,
		Counter: counters,
	}
	return s.store.SaveMetricsToFile(filename, &metrics)
}

func (s *MetricsService) ReadMetricsFromFile(filename string) error {
	loadedData, err := s.store.ReadMetricsFromFile(filename)
	if err != nil {
		return err
	}
	for name, valueCounter := range loadedData.Counter {
		s.store.SaveCounter(name, valueCounter)
	}
	for name, valueGauge := range loadedData.Gauge {
		s.store.SaveGauge(name, valueGauge)
	}
	return err
}

func (s *MetricsService) PingDatabase() error {
	if s.db == nil {
		return errors.New("database is not initialized")
	}

	dbConnect, err := s.db.GetDB()
	if err != nil {
		logger.Log.Error("Can not get DB connection", zap.Error(err))
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := dbConnect.PingContext(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	return nil
}

func (s *MetricsService) SaveBatch(ctx context.Context, metricsSlice []models.Metrics) error {
	for _, v := range metricsSlice {
		switch v.MType {
		case metrics.Counter:
			if v.Delta == nil {
				errMsg := fmt.Sprintf("missing field: delta for counter %q", v.MType)
				logger.Log.Warn("invalid metric", zap.String("error", errMsg))
				continue
			}
			err := s.SaveCounter(v.ID, *v.Delta, ctx)
			if err != nil {
				logger.Log.Error(err.Error())
				return fmt.Errorf("failed to save counter %s: %w", v.ID, err)
			}
		case metrics.Gauge:
			if v.Value == nil {
				errMsg := fmt.Sprintf("missing field: value for gauge %q", v.MType)
				logger.Log.Warn("invalid metric", zap.String("error", errMsg))
			}
			err := s.SaveGauge(v.ID, *v.Value, ctx)
			if err != nil {
				logger.Log.Error(err.Error())
				return fmt.Errorf("failed to save gauge %s: %w", v.ID, err)
			}
		default:
			logger.Log.Warn("unsupported metric type", zap.String("type", v.MType))
		}
	}
	return nil
}
