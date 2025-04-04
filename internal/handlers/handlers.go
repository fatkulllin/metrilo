package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/fatkulllin/metrilo/internal/storage"
	"github.com/go-chi/chi"
)

var memStorage = storage.NewMemoryStorage()

func isLetter(s string) bool {
	return !strings.ContainsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r)
	})
}

func SaveMetrics(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain")
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!!", http.StatusMethodNotAllowed)
		return
	}
	if req.Header.Get("Content-Type") != "text/plain" {
		http.Error(res, "Only Content-Type: text/plain header are allowed!!", http.StatusMethodNotAllowed)
		return
	}
	typeMetric := chi.URLParam(req, "type")
	nameMetric := chi.URLParam(req, "name")
	valueMetric := chi.URLParam(req, "value")
	if !isLetter(nameMetric) {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if typeMetric == "" || nameMetric == "" || valueMetric == "" {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	if typeMetric != "gauge" && typeMetric != "counter" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if typeMetric == "counter" {
		// if !metrics.IsMetricCounterAllowed(nameMetric) {
		// 	res.WriteHeader(http.StatusBadRequest)
		// 	return
		// }
		incrementValue, err := strconv.ParseInt(valueMetric, 10, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		memStorage.AddCounter(nameMetric, incrementValue)
		// storage.AddCounter(nameMetric, incrementValue)
	}
	if typeMetric == "gauge" {
		// if !metrics.IsMetricGaugeAllowed(nameMetric) {
		// 	res.WriteHeader(http.StatusBadRequest)
		// 	return
		// }
		floatValue, err := strconv.ParseFloat(valueMetric, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		// fmt.Println(floatValue)
		// fmt.Println(incrementValue)
		memStorage.SetGauge(nameMetric, floatValue)
		// m.Gauge[nameMetric] = incrementValue
	}

	res.WriteHeader(http.StatusOK)
}

func GetMetric(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain")
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!!", http.StatusMethodNotAllowed)
		return
	}
	if req.Header.Get("Content-Type") != "text/plain" {
		http.Error(res, "Only Content-Type: text/plain header are allowed!!", http.StatusMethodNotAllowed)
		return
	}
	typeMetric := chi.URLParam(req, "type")
	nameMetric := chi.URLParam(req, "name")
	if typeMetric == "" || nameMetric == "" {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	if typeMetric != "gauge" && typeMetric != "counter" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if typeMetric == "counter" {
		// if !metrics.IsMetricCounterAllowed(nameMetric) {
		// 	res.WriteHeader(http.StatusBadRequest)
		// 	return
		// }
		result, err := memStorage.GetCounter(nameMetric)
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		io.WriteString(res, strconv.FormatInt(result, 10))
		// storage.AddCounter(nameMetric, incrementValue)
	}
	if typeMetric == "gauge" {
		// if !metrics.IsMetricGaugeAllowed(nameMetric) {
		// 	res.WriteHeader(http.StatusBadRequest)
		// 	return
		// }
		// fmt.Println(floatValue)
		// fmt.Println(incrementValue)
		result, err := memStorage.GetGauge(nameMetric)
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		io.WriteString(res, strconv.FormatFloat(result, 'f', 2, 64))
		// m.Gauge[nameMetric] = incrementValue
	}

	res.WriteHeader(http.StatusOK)
}

func AllGetMetrics(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html")
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!!", http.StatusMethodNotAllowed)
		return
	}
	if req.Header.Get("Content-Type") != "text/plain" {
		http.Error(res, "Only Content-Type: text/plain header are allowed!!", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintln(res, "<ul>")
	for k, v := range memStorage.Counter {
		fmt.Fprintf(res, "<li>%s: %.v</li>\n", k, v)
	}
	for k, v := range memStorage.Gauge {
		fmt.Fprintf(res, "<li>%s: %v</li>\n", k, v)
	}
	fmt.Fprintln(res, "</ul>")
}
