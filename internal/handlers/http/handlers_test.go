package httphandlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

// Пример для метода (*Handlers).SaveJSONMetrics
func ExampleHandlers_SaveJSONMetrics() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/update/" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		io.Copy(w, r.Body)
	}))
	defer srv.Close()

	jsonBody := []byte(`{"id":"requests","type":"counter","delta":1}`)
	resp, err := http.Post(srv.URL+"/update/", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))

	// Output:
	// {"id":"requests","type":"counter","delta":1}
}

// Пример для метода (*Handlers).GetMetric
func ExampleHandlers_GetMetric() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/value/counter/requests" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("42"))
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/value/counter/requests")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	fmt.Println(string(b))

	// Output:
	// 42
}

// Пример для метода (*Handlers).AllGetMetrics
func ExampleHandlers_AllGetMetrics() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintln(w, "<ul>")
			fmt.Fprintln(w, "<li>requests: 42</li>")
			fmt.Fprintln(w, "<li>cpu: 0.12</li>")
			fmt.Fprintln(w, "</ul>")
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	fmt.Println(string(b))

	// Output:
	// <ul>
	// <li>requests: 42</li>
	// <li>cpu: 0.12</li>
	// </ul>
}
