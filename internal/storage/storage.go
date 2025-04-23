package storage

import (
	"errors"
	"fmt"

	"github.com/fatkulllin/metrilo/internal/logger"
)

type Repositories interface {
	SaveGauge(name string, value float64)
	SaveCounter(name string, value int64)
	GetCounter(name string)
	GetGauge(name string)
}

type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

func NewMemoryStorage() *MemStorage {
	logger.Log.Info("Initializing memory storage...")
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

func (m *MemStorage) SaveCounter(nameMetric string, increment int64) {
	m.Counter[nameMetric] += increment
	fmt.Printf("Save type Counter %+v\n", m)
}

func (m *MemStorage) SaveGauge(nameMetric string, increment float64) {
	m.Gauge[nameMetric] = increment
	fmt.Printf("Save type Gauge %+v\n", m)
}

func (m *MemStorage) GetCounter(nameMetric string) (int64, error) {
	value, exists := m.Counter[nameMetric]
	if !exists {
		return 0, errors.New("metric not found")
	}
	return value, nil
}

func (m *MemStorage) GetGauge(nameMetric string) (float64, error) {
	value, exists := m.Gauge[nameMetric]
	if !exists {
		return 0, errors.New("metric not found")
	}
	return value, nil
}

func (m *MemStorage) GetMetrics() (map[string]float64, map[string]int64) {
	return m.Gauge, m.Counter
}
