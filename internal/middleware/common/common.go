package common

import (
	"net/http"
	"strings"
	"unicode"

	"github.com/go-chi/chi"
)

func SetHeaderTextMiddleware(next http.Handler) http.Handler {
	// получаем Handler приведением типа http.HandlerFunc
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/plain")
		next.ServeHTTP(res, req)
	})
}

func SetHeaderHTMLMiddleware(next http.Handler) http.Handler {
	// получаем Handler приведением типа http.HandlerFunc
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/html")
		next.ServeHTTP(res, req)
	})
}

func isLetter(s string) bool {
	return !strings.ContainsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r)
	})
}

func CheckReqHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// metricType := chi.URLParam(req, "type")
		metricName := chi.URLParam(req, "name")
		// metricValue := chi.URLParam(req, "value")

		if isLetter(metricName) {
			if req.Header.Get("Content-Type") != "text/plain" {
				http.Error(res, "Only Content-Type: text/plain header are allowed!!", http.StatusMethodNotAllowed)
				return
			}
		}
		next.ServeHTTP(res, req)
	})
}

func MethodPostOnlyMiddleware(next http.Handler) http.Handler {
	// получаем Handler приведением типа http.HandlerFunc
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(res, "Only POST requests are allowed!!", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(res, req)
	})
}

func MethodGetOnlyMiddleware(next http.Handler) http.Handler {
	// получаем Handler приведением типа http.HandlerFunc
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(res, "Only GET requests are allowed!!", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(res, req)
	})
}

func ValidateURLParamsMiddleware(next http.Handler) http.Handler {
	// получаем Handler приведением типа http.HandlerFunc
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		typeMetric := chi.URLParam(req, "type")
		nameMetric := chi.URLParam(req, "name")
		valueMetric := chi.URLParam(req, "value")

		if typeMetric == "" || nameMetric == "" || valueMetric == "" {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		if typeMetric != "gauge" && typeMetric != "counter" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		next.ServeHTTP(res, req)
	})
}

func ValidateTypeMetricMiddleware(next http.Handler) http.Handler {
	// получаем Handler приведением типа http.HandlerFunc
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		typeMetric := chi.URLParam(req, "type")

		if typeMetric != "gauge" && typeMetric != "counter" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		next.ServeHTTP(res, req)
	})
}
