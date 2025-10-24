package gprchandlers

import (
	"context"

	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/fatkulllin/metrilo/internal/models"
	service "github.com/fatkulllin/metrilo/internal/service/server"
	proto "github.com/fatkulllin/metrilo/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricsGRPCServer struct {
	proto.UnimplementedMetricsServiceServer
	service *service.MetricsService
}

func NewMetricsGRPCServer(service *service.MetricsService) *MetricsGRPCServer {
	return &MetricsGRPCServer{service: service}
}

func (m *MetricsGRPCServer) UpdateMetrics(ctx context.Context, in *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	logger.Log.Info("Received gRPC UpdateMetrics request")
	if in == nil || len(in.GetMetrics()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty metrics batch")
	}
	metricsSlice := make([]models.Metrics, len(in.GetMetrics()))
	for i, metric := range in.GetMetrics() {
		mm := models.Metrics{
			ID:    metric.GetId(),
			MType: metric.GetType(),
		}

		switch metric.GetType() {
		case "gauge":
			value := metric.GetValue()
			mm.Value = &value
		case "counter":
			delta := metric.GetDelta()
			mm.Delta = &delta
		}

		metricsSlice[i] = mm
	}
	if err := m.service.SaveBatch(ctx, metricsSlice); err != nil {
		logger.Log.Error("failed to save metrics batch", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to save metrics: %v", err)
	}
	logger.Log.Info("Metrics batch saved successfully via gRPC")
	return &proto.UpdateMetricsResponse{}, nil
}
