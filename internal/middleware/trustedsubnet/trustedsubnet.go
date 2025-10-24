package trustedsubnet

import (
	"net"
	"net/http"

	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/fatkulllin/metrilo/internal/trusted"
	"go.uber.org/zap"
)

func TrustedSubnetMiddleware(trustedNet *net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

			ipStr := req.Header.Get("X-Real-IP")

			if err := trusted.ValidateIP(ipStr, trustedNet); err != nil {
				logger.Log.Warn("unauthorized IP", zap.String("IP", ipStr))
				http.Error(res, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(res, req)
		})
	}
}
