package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/fatkulllin/metrilo/internal/metrics"
)

func httpClient() *http.Client {
	client := &http.Client{Timeout: 10 * time.Second}
	return client
}

func main() {
	metrics := metrics.NewMetrics()
	pollInterval := time.NewTicker(time.Duration(2) * time.Second)
	reportInterval := time.NewTicker(time.Duration(10) * time.Second)
	// lastSendMetricsTime := time.Now().Second()
	endpoint := ""
	c := httpClient()

	defer pollInterval.Stop()
	defer reportInterval.Stop()

	for {
		select {
		case <-pollInterval.C:
			metrics.CollectMetrics()
		case <-reportInterval.C:
			go func() {
				fmt.Println("Send Gauge type")
				for k, v := range metrics.Gauge {
					fmt.Printf("Send Gauge type http://localhost:8080/update/gauge/%v/%v\n", k, v)
					endpoint = fmt.Sprintf("http://localhost:8080/update/gauge/%v/%v", k, v)
					sendRequest(c, http.MethodPost, endpoint)
				}
			}()
			go func() {
				fmt.Println("Send Counter type")
				for k, v := range metrics.Counter {
					fmt.Printf("Send Counter type http://localhost:8080/update/counter/%v/%v\n", k, v)
					endpoint = fmt.Sprintf("http://localhost:8080/update/counter/%v/%v", k, v)
					sendRequest(c, http.MethodPost, endpoint)
				}
			}()
		}

	}
}

func sendRequest(client *http.Client, method string, endpoint string) ([]byte, int) {

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer([]byte{}))
	if err != nil {
		log.Fatalf("Error Occurred. %+v", err)
	}
	req.Header.Add("Content-Type", "text/plain")
	response, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request to API endpoint. %+v", err)
	}

	// Close the connection to reuse it
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Couldn't parse response body. %+v", err)
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Request failed with status: %s", response.Status)
	}
	fmt.Printf("Тело ответа: %s\n", body)
	return body, response.StatusCode
}
