package service

import "github.com/fatkulllin/metrilo/internal/storage"

type MetricsService struct {
	store *storage.MemStorage
}

func NewMetricsService(store *storage.MemStorage) *MetricsService {
	return &MetricsService{store: store}
}

func (s *MetricsService) SaveGauge(name string, value float64) {
	s.store.SaveGauge(name, value)
}

func (s *MetricsService) SaveCounter(name string, delta int64) {
	s.store.SaveCounter(name, delta)
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
