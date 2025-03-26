package main

import (
	"fmt"
	"net/http"
	"strconv"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

// функция run будет полезна при инициализации зависимостей сервера перед запуском
func run() error {
	mux := initilizeRoutes()
	return http.ListenAndServe(`:8080`, mux)
}

func initilizeRoutes() *http.ServeMux {

	metrics := &MemStorage{}
	metrics.counter = make(map[string]int64)
	metrics.gauge = make(map[string]float64)

	mux := http.NewServeMux()
	mux.HandleFunc(`/update/{type}/{name}/{value}`, metrics.saveMetrics)
	return mux
}

func (m *MemStorage) saveMetrics(res http.ResponseWriter, req *http.Request) {
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
	}

	if typeMetric != "gauge" && typeMetric != "counter" {
		res.WriteHeader(http.StatusBadRequest)
	}
	if typeMetric == "counter" {
		incrementValue, err := strconv.ParseInt(valueMetric, 10, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
		}
		m.counter[nameMetric] += incrementValue
	}
	if typeMetric == "gauge" {
		incrementValue, err := strconv.ParseFloat(valueMetric, 64)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
		}
		m.gauge[nameMetric] = incrementValue
	}
	fmt.Println(m)
	res.WriteHeader(http.StatusOK)
}
