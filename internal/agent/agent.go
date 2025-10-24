package agent

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	config "github.com/fatkulllin/metrilo/internal/config/agent"
	"github.com/fatkulllin/metrilo/internal/encoder"
	"github.com/fatkulllin/metrilo/internal/gzip"
	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/fatkulllin/metrilo/internal/models"
	service "github.com/fatkulllin/metrilo/internal/service/agent"
	proto "github.com/fatkulllin/metrilo/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Agent struct {
	ServerAddress  string
	ReportInterval int
	PollInterval   int
	Service        *service.MetricsService
	config         *config.Config
	RateLimit      int
	PublicKey      *rsa.PublicKey
}

func NewAgent(svc *service.MetricsService, cfg *config.Config, publicKey *rsa.PublicKey) *Agent {
	logger.Log.Info("Initializing Agent...")
	agent := &Agent{
		ServerAddress:  cfg.ServerAddress,
		ReportInterval: cfg.ReportInterval,
		PollInterval:   cfg.PollInterval,
		Service:        svc,
		config:         cfg,
		RateLimit:      cfg.RateLimit,
		PublicKey:      publicKey,
	}
	logger.Log.Info("Server address", zap.String("address: ", agent.ServerAddress))
	logger.Log.Info("Report Interval:", zap.Int("report interval: ", agent.ReportInterval))
	logger.Log.Info("Poll Interval:", zap.Int("poll interval: ", agent.PollInterval))
	logger.Log.Info("Agent Host IP", zap.String("agent host ip: ", agent.config.AgentHostIP))
	return agent
}

func newHTTPClient() *http.Client {
	client := &http.Client{}
	return client
}

func createGRPCClient(address, hostIP string) (proto.MetricsServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		"dns:///"+address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(ClientIPInterceptor(hostIP)),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}
	client := proto.NewMetricsServiceClient(conn)
	return client, conn, nil
}

func (agent *Agent) Run(ctx context.Context) error {
	pollInterval := time.NewTicker(time.Duration(agent.PollInterval) * time.Second)
	defer pollInterval.Stop()
	reportInterval := time.NewTicker(time.Duration(agent.ReportInterval) * time.Second)
	defer reportInterval.Stop()

	var m sync.RWMutex
	jobs := make(chan []models.Metrics, 10)

	var wg sync.WaitGroup
	workerCount := agent.config.RateLimit
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			agent.worker(ctx, id, agent.PublicKey, jobs)
		}(i)
	}

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Agent shutting down...")

			close(jobs)
			wg.Wait()
			logger.Log.Info("Agent stopped gracefully")
			return nil

		case <-pollInterval.C:
			agent.Service.CollectMetrics(&m)
			agent.Service.CollectGopsutilMetrics(&m)

		case <-reportInterval.C:
			m.RLock()
			collected := agent.buildMetrics()
			m.RUnlock()

			batchSize := 10
			for i := 0; i < len(collected); i += batchSize {
				end := min(i+batchSize, len(collected))
				batch := collected[i:end]

				jobs <- batch
			}
		}
	}
}

func (agent *Agent) buildMetrics() []models.Metrics {
	metrics := make([]models.Metrics, 0)
	for k, v := range agent.Service.GetMetrics().Gauge {
		metrics = append(metrics, models.Metrics{
			ID:    k,
			MType: "gauge",
			Value: &v})
	}
	for k, v := range agent.Service.GetMetrics().Counter {
		metrics = append(metrics, models.Metrics{
			ID:    k,
			MType: "counter",
			Delta: &v})
	}
	return metrics
}
func sendToServerGRPC(ctx context.Context, client proto.MetricsServiceClient, batch []models.Metrics) error {
	metricsProto := make([]*proto.Metric, len(batch))
	for i, metric := range batch {
		protoMetric := &proto.Metric{}
		protoMetric.SetId(metric.ID)
		protoMetric.SetType(metric.MType)
		if metric.Value != nil {
			protoMetric.SetValue(*metric.Value)
		}
		if metric.Delta != nil {
			protoMetric.SetDelta(*metric.Delta)
		}
		metricsProto[i] = protoMetric
	}
	req := &proto.UpdateMetricsRequest{}
	req.SetMetrics(metricsProto)
	_, err := client.UpdateMetrics(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send metrics via gRPC: %w", err)
	}
	logger.Log.Info("Sent metrics batch via gRPC", zap.Int("count", len(batch)))
	return nil
}

func (agent *Agent) worker(ctx context.Context, id int, publicKey *rsa.PublicKey, jobs <-chan []models.Metrics) {
	endpoint := fmt.Sprintf("http://%v/updates/", agent.ServerAddress)
	client := newHTTPClient()
	label := []byte("agent")
	hostIP := agent.config.AgentHostIP

	var grpcClient proto.MetricsServiceClient
	var grpcConn *grpc.ClientConn
	var err error

	if agent.config.GRPCAddress != "" {
		grpcClient, grpcConn, err = createGRPCClient(agent.config.GRPCAddress, hostIP)
		if err != nil {
			logger.Log.Error("failed to connect gRPC", zap.Error(err))
			return
		}
		defer grpcConn.Close()
	}

	for batch := range jobs {
		func() {
			reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			var reqBody, gzipped, ciphertext []byte
			var intErr error

			if agent.config.GRPCAddress != "" {
				intErr = sendToServerGRPC(reqCtx, grpcClient, batch)
			} else {
				reqBody, intErr = json.Marshal(batch)
				if intErr != nil {
					logger.Log.Error("marshal batch", zap.Error(intErr))
					return
				}

				gzipped, intErr = gzip.GzipCompress(reqBody)
				if intErr != nil {
					logger.Log.Error("gzip batch", zap.Error(intErr))
					return
				}

				ciphertext, intErr = encoder.BuildPacket(publicKey, gzipped, label)
				if intErr != nil {
					logger.Log.Error("build encode packet error", zap.Error(intErr))
					return
				}

				intErr = agent.Service.SendToServerWithContext(
					reqCtx,
					client,
					http.MethodPost,
					endpoint,
					ciphertext,
					agent.config.WasKeySet,
					[]byte(agent.config.Key),
					hostIP,
				)
			}

			if intErr != nil {
				logger.Log.Error("Worker failed to send batch",
					zap.Int("workerID", id),
					zap.Error(intErr))
			} else {
				logger.Log.Info("Worker sent batch successfully",
					zap.Int("workerID", id),
					zap.Int("batchSize", len(batch)))
			}
		}()
	}
}
