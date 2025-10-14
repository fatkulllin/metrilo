package server

import (
	"context"
	"fmt"
	"net"

	"github.com/fatkulllin/metrilo/internal/logger"
	grpcmw "github.com/fatkulllin/metrilo/internal/middleware/grpc"
	proto "github.com/fatkulllin/metrilo/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (server *Server) StartGRPC(ctx context.Context) error {
	listen, err := net.Listen("tcp", server.config.GRPCAddress)
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC address: %w", err)
	}

	var trustedNet *net.IPNet
	if server.config.TrustedSubnet != "" {
		_, trustedNet, err = net.ParseCIDR(server.config.TrustedSubnet)
		if err != nil {
			logger.Log.Error("invalid trusted subnet", zap.Error(err))
		}
		logger.Log.Info("trusted subnetd", zap.String("cidr", trustedNet.String()))
	}

	serverGRPC := grpc.NewServer(
		grpc.UnaryInterceptor(grpcmw.TrustedSubnetInterceptor(trustedNet)),
	)
	proto.RegisterMetricsServiceServer(serverGRPC, server.grpchandlers)

	go func() {
		<-ctx.Done()
		logger.Log.Info("Shutting down gRPC server gracefully")
		serverGRPC.GracefulStop()
	}()

	if err := serverGRPC.Serve(listen); err != nil {
		if err == grpc.ErrServerStopped {
			logger.Log.Info("gRPC server stopped gracefully")
			return nil
		}
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}
	return nil
}
