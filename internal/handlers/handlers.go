package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	service "github.com/fatkulllin/metrilo/internal/service/server"
	"github.com/go-chi/chi"
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
		h.service.SaveCounter(nameMetric, incrementValue)
	case "gauge":
		floatValue, err := strconv.ParseFloat(valueMetric, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		h.service.SaveGauge(nameMetric, floatValue)

	default:
		http.Error(res, "Unknown type", http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusOK)
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

func (h *Handlers) AllGetMetrics(res http.ResponseWriter, req *http.Request) {
	metricsCounter, metricsGauge := h.service.GetMetrics()

	fmt.Fprintln(res, "<ul>")
	for k, v := range metricsCounter {
		fmt.Fprintf(res, "<li>%s: %.v</li>\n", k, v)
	}

	for k, v := range metricsGauge {
		fmt.Fprintf(res, "<li>%s: %v</li>\n", k, v)
	}

	fmt.Fprintln(res, "</ul>")
}
