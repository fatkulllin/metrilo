package grpcmw

import (
	"context"
	"net"
	"strings"

	"github.com/fatkulllin/metrilo/internal/logger"
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
		if trustedNet == nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Log.Warn("no metadata in context")
			return nil, status.Error(codes.PermissionDenied, "missing metadata")
		}

		values := md.Get("x-real-ip")
		if len(values) == 0 {
			return nil, status.Error(codes.PermissionDenied, "missing X-Real-IP header")
		}

		ipStr := strings.TrimSpace(values[0])
		ip := net.ParseIP(ipStr)
		if ip == nil {
			logger.Log.Warn("invalid X-Real-IP", zap.String("ip", ipStr))
			return nil, status.Error(codes.PermissionDenied, "invalid IP address")
		}

		if !trustedNet.Contains(ip) {
			logger.Log.Warn("unauthorized IP", zap.String("ip", ipStr))
			return nil, status.Error(codes.PermissionDenied, "IP not allowed")
		}

		return handler(ctx, req)
	}
}
