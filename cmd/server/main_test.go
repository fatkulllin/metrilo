package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fatkulllin/metrilo/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, headers map[string]string) (*http.Response, string) {
	fmt.Println(ts.URL + path)
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	fmt.Println(string(respBody))
	require.NoError(t, err)

	return resp, string(respBody)
}
func TestRouter(t *testing.T) {
	address := "localhost:8080"
	server := server.NewServer(&address)
	router := server.Router()
	ts := httptest.NewServer(router)
	defer ts.Close()

	var testTable = []struct {
		url     string
		method  string
		status  int
		want    string
		headers map[string]string
	}{
		{"/update/counter/CountereMetric/1", "POST", http.StatusOK, "", map[string]string{"Content-Type": "text/plain"}},
		{"/update/gauge/GaugeMetric/1", "POST", http.StatusOK, "", map[string]string{"Content-Type": "text/plain"}},
		{"/", "GET", http.StatusOK, "<ul>\n<li>CountereMetric: 1</li>\n<li>GaugeMetric: 1</li>\n</ul>\n", map[string]string{"Content-Type": "text/plain"}},
		// // проверим на ошибочный запрос
		{"/update/gauge/PollInterval/1", "POST", http.StatusMethodNotAllowed, "Only Content-Type: text/plain header are allowed!!\n", map[string]string{"Content-Type": "text/json"}},
		{"/update/gauge/PollInterval/1", "GET", http.StatusMethodNotAllowed, "", map[string]string{"Content-Type": "text/plain"}},
	}
	for _, v := range testTable {
		resp, getBody := testRequest(t, ts, v.method, v.url, v.headers)
		defer resp.Body.Close()
		assert.Equal(t, v.status, resp.StatusCode)
		assert.Equal(t, v.want, getBody)
	}
}
