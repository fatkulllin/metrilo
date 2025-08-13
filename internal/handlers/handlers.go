// Package handlers содержит HTTP-обработчики (хендлеры) для работы с сервисом сбора метрик.
// Хендлеры принимают и обрабатывают HTTP-запросы, валидируют входные данные,
// вызывают соответствующие методы бизнес-логики и формируют HTTP-ответы
package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/fatkulllin/metrilo/internal/metrics"
	"github.com/fatkulllin/metrilo/internal/models"
	service "github.com/fatkulllin/metrilo/internal/service/server"
)

// Handlers — структура, объединяющая HTTP-хендлеры для работы с метриками.
// Использует MetricsService для сохранения, получения и обновления значений метрик.
type Handlers struct {
	service *service.MetricsService
}

// NewHandlers создаёт новый экземпляр Handlers.
// Принимает MetricsService, через который хендлеры работают с данными.
func NewHandlers(service *service.MetricsService) *Handlers {
	return &Handlers{service: service}
}

// SaveMetrics обрабатывает HTTP-запрос для сохранения метрики через путь вида
// /update/{type}/{name}/{value}. Поддерживает типы counter и gauge.
// В случае ошибки валидации возвращает HTTP 400.
func (h *Handlers) SaveMetrics(res http.ResponseWriter, req *http.Request) {
	typeMetric := chi.URLParam(req, "type")
	nameMetric := chi.URLParam(req, "name")
	valueMetric := chi.URLParam(req, "value")

	switch typeMetric {
	case metrics.Counter:
		incrementValue, err := strconv.ParseInt(valueMetric, 10, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		h.service.SaveCounter(nameMetric, incrementValue, req.Context())
	case metrics.Gauge:
		floatValue, err := strconv.ParseFloat(valueMetric, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := h.service.SaveGauge(nameMetric, floatValue, req.Context()); err != nil {
			res.Write([]byte(err.Error()))
			res.WriteHeader(http.StatusBadRequest)
			return
		}

	default:
		http.Error(res, "Unknown type", http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusOK)
}

// SaveJSONMetrics обрабатывает HTTP-запрос с JSON-телом для сохранения метрики.
// Ожидает Content-Type: application/json и структуру Metrics в теле запроса.
// Возвращает сохранённую метрику в JSON-формате.
func (h *Handlers) SaveJSONMetrics(res http.ResponseWriter, req *http.Request) {
	logger.Log.Info("Request:",
		zap.String("method", req.Method),
		zap.String("url", req.URL.String()),
	)

	var r models.Metrics
	logger.Log.Info("decoding request")

	res.Header().Set("Content-Type", "application/json")
	decode := json.NewDecoder(req.Body)
	if err := decode.Decode(&r); err != nil {
		logger.Log.Error("cannot decode request JSON body", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	logger.Log.Info("parsed request", zap.Any("request", r))
	typeMetric := r.MType
	nameMetric := r.ID

	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Only Content-Type: application/json header are allowed!!", http.StatusMethodNotAllowed)
		return
	}
	if r.ID == "" || r.MType == "" {
		http.Error(res, "missing fields", http.StatusBadRequest)
		return
	}
	switch typeMetric {
	case metrics.Counter:
		if r.Delta == nil {
			http.Error(res, "missing required field: delta for counter", http.StatusBadRequest)
			return
		}
		valueMetric := *r.Delta
		h.service.SaveCounter(nameMetric, valueMetric, req.Context())
		resp, err := json.Marshal(models.Metrics{
			ID:    nameMetric,
			MType: metrics.Counter,
			Delta: &valueMetric,
		})
		if err != nil {
			logger.Log.Error(err.Error())
		}
		res.Write(resp)
	case metrics.Gauge:
		if r.Value == nil {
			http.Error(res, "missing required field: value for counter", http.StatusBadRequest)
			return
		}
		valueMetric := *r.Value
		h.service.SaveGauge(nameMetric, valueMetric, req.Context())
		resp, err := json.Marshal(models.Metrics{
			ID:    nameMetric,
			MType: metrics.Gauge,
			Value: &valueMetric,
		})
		if err != nil {
			logger.Log.Error(err.Error())
		}
		res.Write(resp)
	default:
		http.Error(res, "Unknown type", http.StatusBadRequest)
		return
	}
}

// GetMetric возвращает значение метрики в текстовом виде по её типу и имени.
// Параметры передаются через путь /value/{type}/{name}.
func (h *Handlers) GetMetric(res http.ResponseWriter, req *http.Request) {
	typeMetric := chi.URLParam(req, "type")
	nameMetric := chi.URLParam(req, "name")

	switch typeMetric {

	case metrics.Counter:
		result, err := h.service.GetCounter(nameMetric, req.Context())
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		io.WriteString(res, strconv.FormatInt(result, 10))

	case metrics.Gauge:
		result, err := h.service.GetGauge(nameMetric, req.Context())
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		io.WriteString(res, strconv.FormatFloat(result, 'f', -1, 64))

	default:
		http.Error(res, "Unknown type", http.StatusBadRequest)
		return
	}
}

// GetMetricJSON возвращает значение метрики в JSON-формате.
// Ожидает JSON-запрос с полями ID и MType.
func (h *Handlers) GetMetricJSON(res http.ResponseWriter, req *http.Request) {
	var r models.Metrics
	logger.Log.Info("decoding request")

	res.Header().Set("Content-Type", "application/json")
	decode := json.NewDecoder(req.Body)
	if err := decode.Decode(&r); err != nil {
		logger.Log.Info("cannot decode request JSON body", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	logger.Log.Info("parsed request", zap.Any("request", r))
	typeMetric := r.MType
	nameMetric := r.ID

	switch typeMetric {

	case metrics.Counter:
		result, err := h.service.GetCounter(nameMetric, req.Context())
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		if err := json.NewEncoder(res).Encode(models.Metrics{
			ID:    nameMetric,
			MType: typeMetric,
			Delta: &result,
		}); err != nil {
			logger.Log.Error("failed to encode metrics")
			http.Error(res, "failed to encode metrics", http.StatusInternalServerError)
			return
		}

	case metrics.Gauge:
		result, err := h.service.GetGauge(nameMetric, req.Context())
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		if err := json.NewEncoder(res).Encode(models.Metrics{
			ID:    nameMetric,
			MType: typeMetric,
			Value: &result,
		}); err != nil {
			logger.Log.Error("failed to encode metrics")
			http.Error(res, "failed to encode metrics", http.StatusInternalServerError)
			return
		}

	default:
		http.Error(res, "Unknown type", http.StatusBadRequest)
		return
	}
}

// AllGetMetrics выводит HTML-страницу со списком всех сохранённых метрик.
func (h *Handlers) AllGetMetrics(res http.ResponseWriter, req *http.Request) {
	metricsGauge, metricsCounter := h.service.GetMetrics()

	fmt.Fprintln(res, "<ul>")
	for k, v := range metricsCounter {
		fmt.Fprintf(res, "<li>%s: %.v</li>\n", k, v)
	}

	for k, v := range metricsGauge {
		fmt.Fprintf(res, "<li>%s: %v</li>\n", k, v)
	}

	fmt.Fprintln(res, "</ul>")
}

// PingDatabase выполняет проверку доступности базы данных.
// Возвращает HTTP 200, если соединение установлено, иначе HTTP 500.
func (h *Handlers) PingDatabase(res http.ResponseWriter, req *http.Request) {
	err := h.service.PingDatabase()
	if err != nil {
		logger.Log.Error("database is not connected", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}

// UpdateMetrics принимает JSON-массив метрик и обновляет их значения в хранилище.
// Для каждой метрики проверяет наличие обязательных полей (Delta для counter, Value для gauge).
// При ошибке сохранения возвращает HTTP 500.
func (h *Handlers) UpdateMetrics(res http.ResponseWriter, req *http.Request) {
	logger.Log.Info("Request:",
		zap.String("method", req.Method),
		zap.String("url", req.URL.String()),
	)

	metricsSlice := make([]models.Metrics, 0)

	logger.Log.Info("decoding request")

	decode := json.NewDecoder(req.Body)

	if err := decode.Decode(&metricsSlice); err != nil {
		logger.Log.Error("cannot decode request JSON body", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Log.Info("parsed request", zap.Any("request", metricsSlice))

	for _, v := range metricsSlice {
		switch v.MType {
		case metrics.Counter:
			if v.Delta == nil {
				errMsg := fmt.Sprintf("missing field: delta for counter %q", v.MType)
				logger.Log.Warn("invalid request", zap.String("error", errMsg))
				http.Error(res, errMsg, http.StatusBadRequest)
			}
			err := h.service.SaveCounter(v.ID, *v.Delta, req.Context())
			if err != nil {
				logger.Log.Error(err.Error())
				http.Error(res, "DB connect is not available", http.StatusInternalServerError)
				return
			}
		case metrics.Gauge:
			if v.Value == nil {
				errMsg := fmt.Sprintf("missing field: value for gauge %q", v.MType)
				logger.Log.Warn("invalid request", zap.String("error", errMsg))
				http.Error(res, errMsg, http.StatusBadRequest)
			}
			err := h.service.SaveGauge(v.ID, *v.Value, req.Context())
			if err != nil {
				logger.Log.Error(err.Error())
				http.Error(res, "DB connect is not available", http.StatusInternalServerError)
				return
			}
		default:
			errMsg := fmt.Sprintf("bad request: unsupported metric type %q", v.MType)
			logger.Log.Warn("invalid request", zap.String("error", errMsg))
			http.Error(res, errMsg, http.StatusBadRequest)
		}
	}
}
