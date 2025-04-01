package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/fatkulllin/metrilo/internal/storage"
)

var memStorage = storage.NewMemoryStorage()

func SaveMetrics(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!!", http.StatusMethodNotAllowed)
		return
	}
	if req.Header.Get("Content-Type") != "text/plain" {
		http.Error(res, "Only Content-Type: text/plain header are allowed!!", http.StatusMethodNotAllowed)
		return
	}
	typeMetric := req.PathValue("type")
	nameMetric := req.PathValue("name")
	valueMetric := req.PathValue("value")

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
		return
	}
	if typeMetric == "gauge" {
		// if !metrics.IsMetricGaugeAllowed(nameMetric) {
		// 	res.WriteHeader(http.StatusBadRequest)
		// 	return
		// }
		if !strings.Contains(valueMetric, ".") {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		floatValue, err := strconv.ParseFloat(valueMetric, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		// fmt.Println(floatValue)
		// fmt.Println(incrementValue)
		memStorage.SetGauge(nameMetric, floatValue)
		// m.Gauge[nameMetric] = incrementValue
		return
	}

	res.WriteHeader(http.StatusOK)

}
