package common

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
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
		metricName := chi.URLParam(req, "name")
		if isLetter(metricName) && req.Header.Get("Content-Type") != "text/plain" {
			http.Error(res, "Only Content-Type: text/plain header are allowed!!", http.StatusMethodNotAllowed)
			return
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

func DecodeMsg(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

		secretkey := []byte("secretkey")
		bodyBytes, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		encodeHeader := req.Header.Get("HashSHA256")
		data, err := hex.DecodeString(encodeHeader)
		if err != nil {
			panic(err)
		}
		h := hmac.New(sha256.New, secretkey)
		h.Write(bodyBytes)
		sign := h.Sum(nil)
		fmt.Printf("%x\n", sign)

		if hmac.Equal(data, sign) {
			fmt.Println("Подпись подлинная")
		} else {
			fmt.Println("Подпись неверна. Где-то ошибка")
		}
		next.ServeHTTP(res, req)
	})
}
