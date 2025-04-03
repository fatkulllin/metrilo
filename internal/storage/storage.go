package storage

import (
	"errors"
	"fmt"
)

type Repositories interface {
	SetGauge(name string, value float64)
	AddCounter(name string, value int64)
}

type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

func NewMemoryStorage() *MemStorage {
	fmt.Println("Initializing memory storage...")
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

func (m *MemStorage) AddCounter(nameMetric string, increment int64) {
	m.Counter[nameMetric] += increment
	fmt.Printf("Save type Counter %+v\n", m)
}

func (m *MemStorage) SetGauge(nameMetric string, increment float64) {
	m.Gauge[nameMetric] = increment
	fmt.Printf("Save type Gauge %+v\n", m)
}

func (m *MemStorage) GetGauge(nameMetric string) (float64, error) {
	value, exists := m.Gauge[nameMetric]
	if !exists {
		return 0, errors.New("metric not found")
	}
	return value, nil
}

func (m *MemStorage) GetCounter(nameMetric string) (int64, error) {
	value, exists := m.Counter[nameMetric]
	if !exists {
		return 0, errors.New("metric not found")
	}
	return value, nil
}
