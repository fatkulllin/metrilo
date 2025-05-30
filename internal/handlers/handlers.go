package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/fatkulllin/metrilo/internal/models"
	service "github.com/fatkulllin/metrilo/internal/service/server"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

type Handlers struct {
	service *service.MetricsService
}

func NewHandlers(service *service.MetricsService) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) SaveMetrics(res http.ResponseWriter, req *http.Request) {
	typeMetric := chi.URLParam(req, "type")
	nameMetric := chi.URLParam(req, "name")
	valueMetric := chi.URLParam(req, "value")

	switch typeMetric {
	case "counter":
		incrementValue, err := strconv.ParseInt(valueMetric, 10, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		h.service.SaveCounter(nameMetric, incrementValue, req.Context())
	case "gauge":
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
	case "counter":
		if r.Delta == nil {
			http.Error(res, "missing required field: delta for counter", http.StatusBadRequest)
			return
		}
		valueMetric := *r.Delta
		h.service.SaveCounter(nameMetric, valueMetric, req.Context())
		resp, err := json.Marshal(models.Metrics{
			ID:    nameMetric,
			MType: "counter",
			Delta: &valueMetric,
		})
		if err != nil {
			logger.Log.Error(err.Error())
		}
		res.Write(resp)
	case "gauge":
		if r.Value == nil {
			http.Error(res, "missing required field: value for counter", http.StatusBadRequest)
			return
		}
		valueMetric := *r.Value
		h.service.SaveGauge(nameMetric, valueMetric, req.Context())
		resp, err := json.Marshal(models.Metrics{
			ID:    nameMetric,
			MType: "gauge",
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

func (h *Handlers) GetMetric(res http.ResponseWriter, req *http.Request) {
	typeMetric := chi.URLParam(req, "type")
	nameMetric := chi.URLParam(req, "name")

	switch typeMetric {

	case "counter":
		result, err := h.service.GetCounter(nameMetric)
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		io.WriteString(res, strconv.FormatInt(result, 10))

	case "gauge":
		result, err := h.service.GetGauge(nameMetric)
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

	case "counter":
		result, err := h.service.GetCounter(nameMetric)
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

	case "gauge":
		result, err := h.service.GetGauge(nameMetric)
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

func (h *Handlers) PingDatabase(res http.ResponseWriter, req *http.Request) {
	err := h.service.PingDatabase()
	if err != nil {
		logger.Log.Error("database is not connected", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}
