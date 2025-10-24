package trusted

import (
	"fmt"
	"net"
	"strings"
)

func ValidateIP(ipStr string, trustedNet *net.IPNet) error {

	if trustedNet == nil {
		return nil
	}

	ipStr = strings.TrimSpace(ipStr)
	if ipStr == "" {
		return fmt.Errorf("missing X-Real-IP header")
	}

	ip := net.ParseIP(ipStr)

	if ip == nil {
		return fmt.Errorf("invalid IP(%s) in X-Real-IP", ip)
	}

	if !trustedNet.Contains(ip) {
		return fmt.Errorf("IP(%s) is not allowed", ip)
	}
	return nil
}
