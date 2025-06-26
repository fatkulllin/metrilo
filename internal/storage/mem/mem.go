package mem

import (
	"fmt"

	"github.com/fatkulllin/metrilo/internal/logger"
	"go.uber.org/zap"
)

type MemStorage struct {
	Gauge   map[string]float64 `json:"Gauge"`
	Counter map[string]int64   `json:"Counter"`
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
	logger.Log.Info("Save type Counter", zap.String("name: ", nameMetric), zap.Int64("value: ", increment))
}

func (m *MemStorage) SaveGauge(nameMetric string, increment float64) {
	m.Gauge[nameMetric] = increment
	logger.Log.Info("Save type Gauge", zap.String("name: ", nameMetric), zap.Float64("value: ", increment))
}

func (m *MemStorage) GetCounter(nameMetric string) {
	fmt.Println("Counter")
}

func (m *MemStorage) GetGauge(nameMetric string) {
	fmt.Println("Gauge")
}
