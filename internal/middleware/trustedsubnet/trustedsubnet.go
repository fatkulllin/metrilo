package trustedsubnet

import (
	"net"
	"net/http"
)

func TrustedSubnetMiddleware(trustedNet *net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

			if trustedNet == nil {
				next.ServeHTTP(res, req)
				return
			}

			ipStr := req.Header.Get("X-Real-IP")
			if ipStr == "" {
				http.Error(res, "missing X-Real-IP header", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(ipStr)
			if ip == nil {
				http.Error(res, "invalid IP in X-Real-IP", http.StatusForbidden)
				return
			}

			if !trustedNet.Contains(ip) {
				http.Error(res, "forbidden", http.StatusForbidden)
				return
			}
		})

	}
}
