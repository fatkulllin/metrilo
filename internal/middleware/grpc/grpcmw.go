package grpcmw

import (
	"context"
	"net"

	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/fatkulllin/metrilo/internal/trusted"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TrustedSubnetInterceptor(trustedNet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		md, _ := metadata.FromIncomingContext(ctx)

		values := md.Get("x-real-ip")

		if err := trusted.ValidateIP(values[0], trustedNet); err != nil {
			logger.Log.Warn("unauthorized IP", zap.String("IP", values[0]))
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}

		return handler(ctx, req)
	}
}
