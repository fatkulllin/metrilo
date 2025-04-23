package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	config "github.com/fatkulllin/metrilo/internal/config/agent"
	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/fatkulllin/metrilo/internal/models"
	service "github.com/fatkulllin/metrilo/internal/service/agent"
)

type Agent struct {
	ServerAddress  string
	ReportInterval int
	PollInterval   int
	Service        *service.MetricsService
}

func NewAgent(svc *service.MetricsService, cfg *config.Config) *Agent {
	log.Println("Initializing Agent...")
	agent := &Agent{
		ServerAddress:  cfg.ServerAddress,
		ReportInterval: cfg.ReportInterval,
		PollInterval:   cfg.PollInterval,
		Service:        svc,
	}
	log.Println("Server Address:", agent.ServerAddress)
	log.Println("Report Interval:", agent.ReportInterval)
	log.Println("Poll Interval:", agent.PollInterval)
	return agent
}

func newHTTPClient() *http.Client {
	client := &http.Client{}
	return client
}

func (agent *Agent) Run() {
	pollInterval := time.NewTicker(time.Duration(agent.PollInterval) * time.Second)
	defer pollInterval.Stop()
	reportInterval := time.NewTicker(time.Duration(agent.ReportInterval) * time.Second)
	defer reportInterval.Stop()
	endpoint := fmt.Sprintf("http://%v/update/", agent.ServerAddress)
	client := newHTTPClient()

	for {
		select {
		case <-pollInterval.C:
			agent.Service.CollectMetrics()
		case <-reportInterval.C:
			fmt.Println("Send metrics")
			go func() {
				for k, v := range agent.Service.GetMetrics().Gauge {
					fmt.Printf("Send Gauge type http://%v/update/ key: %v value:%v\n", agent.ServerAddress, k, v)
					reqBody, err := json.Marshal(models.Metrics{
						ID:    k,
						MType: "gauge",
						Value: &v,
					})
					if err != nil {
						logger.Log.Error(err.Error())
					}
					agent.Service.SendToServer(client, http.MethodPost, endpoint, reqBody)
				}
			}()
			go func() {
				for k, v := range agent.Service.GetMetrics().Counter {
					fmt.Printf("Send Counter type http://%v/update/ key: %v value:%v\n", agent.ServerAddress, k, v)
					reqBody, err := json.Marshal(models.Metrics{
						ID:    k,
						MType: "counter",
						Delta: &v,
					})
					if err != nil {
						logger.Log.Error(err.Error())
					}
					agent.Service.SendToServer(client, http.MethodPost, endpoint, reqBody)
				}
			}()
		}

	}
}
