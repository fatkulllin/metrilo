package service

import (
	"github.com/fatkulllin/metrilo/internal/storage"
)

type MetricsService struct {
	store           *storage.MemStorage
	storeInterval   int
	fileStoragePath string
}

func NewMetricsService(store *storage.MemStorage, storeInterval int, fileStoragePath string) *MetricsService {
	return &MetricsService{store: store, storeInterval: storeInterval, fileStoragePath: fileStoragePath}
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
