package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fatkulllin/metrilo/internal/database"
	"github.com/fatkulllin/metrilo/internal/storage"
)

type MetricsService struct {
	store           *storage.MemStorage
	storeInterval   int
	fileStoragePath string
	db              *database.Database
}

func NewMetricsService(store *storage.MemStorage, storeInterval int, fileStoragePath string, db *database.Database) *MetricsService {
	return &MetricsService{store: store, storeInterval: storeInterval, fileStoragePath: fileStoragePath, db: db}
}

func (s *MetricsService) SaveGauge(name string, value float64) {
	s.store.SaveGauge(name, value)
	if s.storeInterval == 0 {
		_ = s.SaveMetricsToFile(s.fileStoragePath)
	}
}

func (s *MetricsService) SaveCounter(name string, delta int64) {
	s.store.SaveCounter(name, delta)
	if s.storeInterval == 0 {
		_ = s.SaveMetricsToFile(s.fileStoragePath)
	}
}

func (s *MetricsService) GetCounter(name string) (int64, error) {
	return s.store.GetCounter(name)
}

func (s *MetricsService) GetGauge(name string) (float64, error) {
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

	db := s.db.GetDB()
	if db == nil {
		return errors.New("database is not connected")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	return nil
}
