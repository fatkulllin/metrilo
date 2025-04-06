package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/fatkulllin/metrilo/internal/metrics"
	"github.com/spf13/pflag"
)

func HTTPClient() *http.Client {
	client := &http.Client{}
	return client
}

type Agent struct {
	ServerAddress  *string
	ReportInterval *int
	PollInterval   *int
}

func (agent *Agent) initFlags() {
	address := pflag.StringP("address", "a", "localhost:8080", "server address")
	reportInterval := pflag.IntP("reportInterval", "r", 10, "frequency send")
	pollInterval := pflag.IntP("pollInterval", "p", 2, "refresh metric")

	pflag.Parse()
	fmt.Println("Server Address:", *address)
	fmt.Println("Report Interval:", *reportInterval)
	fmt.Println("Poll Interval:", *pollInterval)
	agent.ServerAddress = address
	agent.ReportInterval = reportInterval
	agent.PollInterval = pollInterval
}
func NewAgent() *Agent {
	fmt.Println("Initializing Agent...")
	agent := &Agent{}
	agent.initFlags()
	return agent
}

func main() {
	agent := NewAgent()
	metrics := metrics.NewMetrics()
	pollInterval := time.NewTicker(time.Duration(*agent.PollInterval) * time.Second)
	reportInterval := time.NewTicker(time.Duration(*agent.ReportInterval) * time.Second)
	endpoint := ""
	c := HTTPClient()

	defer pollInterval.Stop()
	defer reportInterval.Stop()

	for {
		select {
		case <-pollInterval.C:
			metrics.CollectMetrics()
		case <-reportInterval.C:
			fmt.Println("Send metrics")
			go func() {
				for k, v := range metrics.Gauge {
					fmt.Printf("Send Gauge type http://%v/update/gauge/%v/%v\n", *agent.ServerAddress, k, v)
					endpoint = fmt.Sprintf("http://%v/update/gauge/%v/%v", *agent.ServerAddress, k, v)
					SendRequest(c, http.MethodPost, endpoint)
				}
			}()
			go func() {
				for k, v := range metrics.Counter {
					fmt.Printf("Send Gauge type http://%v/update/counter/%v/%v\n", *agent.ServerAddress, k, v)
					endpoint = fmt.Sprintf("http://%v/update/counter/%v/%v", *agent.ServerAddress, k, v)
					SendRequest(c, http.MethodPost, endpoint)
				}
			}()
		}

	}
}

func SendRequest(client *http.Client, method string, endpoint string) ([]byte, int) {

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
