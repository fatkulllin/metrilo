package storage

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/fatkulllin/metrilo/internal/logger"
	"go.uber.org/zap"
)

type Repositories interface {
	SaveGauge(name string, value float64)
	SaveCounter(name string, value int64)
	GetCounter(name string)
	GetGauge(name string)
}

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

func (m *MemStorage) GetGaugeDB(dbConnect *sql.DB, nameMetric string, ctx context.Context) (float64, error) {

	row := dbConnect.QueryRowContext(ctx, "SELECT value FROM gauge WHERE name = $1", nameMetric)

	var result float64
	err := row.Scan(&result)
	if err != nil {
		logger.Log.Error("Cannot scan query", zap.Error(err), zap.String("name metric", nameMetric))
	}
	return result, nil
}

func (m *MemStorage) GetCounterDB(dbConnect *sql.DB, nameMetric string, ctx context.Context) (int64, error) {
	row := dbConnect.QueryRowContext(ctx, "SELECT value FROM counter WHERE name = $1", nameMetric)

	var result int64
	err := row.Scan(&result)
	if err != nil {
		logger.Log.Error("Cannot scan query", zap.Error(err), zap.String("name metric", nameMetric))
	}
	return result, nil
}

func (m *MemStorage) GetMetrics() (map[string]float64, map[string]int64) {
	return m.Gauge, m.Counter
}

func (m *MemStorage) SaveMetricsToFile(filename string, metrics *MemStorage) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	data, err := json.MarshalIndent(*metrics, "", " ")
	if err != nil {
		return err
	}
	// записываем событие в буфер
	if _, err := writer.Write(data); err != nil {
		return err
	}

	// добавляем перенос строки
	if err := writer.WriteByte('\n'); err != nil {
		return err
	}

	// записываем буфер в файл
	return writer.Flush()
}

func (m *MemStorage) ReadMetricsFromFile(filename string) (*MemStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	metrics := MemStorage{}

	err = json.Unmarshal(data, &metrics)
	if err != nil {
		return nil, err
	}

	return &metrics, err
}

func (m *MemStorage) SaveGaugeToDB(dbConnect *sql.DB, nameMetric string, increment float64, ctx context.Context) error {
	tx, err := dbConnect.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO gauge (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value;")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, nameMetric, increment)
	if err != nil {
		return fmt.Errorf("storage failed to upsert gauge name: %s error: %s", nameMetric, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("storage failed to commit transaction gauge name: %s error: %s", nameMetric, err)
	}

	return nil
}

func (m *MemStorage) SaveCounterToDB(dbConnect *sql.DB, nameMetric string, increment int64, ctx context.Context) error {

	row := dbConnect.QueryRowContext(ctx, "SELECT value FROM counter WHERE name = $1", nameMetric)

	var current int64

	err := row.Scan(&current)
	if err != nil {
		if err == sql.ErrNoRows {
			current = 0
		} else {
			return fmt.Errorf("error get value counter name: %s error: %s", nameMetric, err)
		}
	}

	result := increment + current

	tx, err := dbConnect.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("storage failed start transaction: %s error: %s", nameMetric, err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO counter (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value;")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, nameMetric, result)
	if err != nil {
		return fmt.Errorf("storage failed to upsert counter name: %s error: %s", nameMetric, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("storage failed to commit transaction counter name: %s error: %s", nameMetric, err)
	}
	return nil
}
