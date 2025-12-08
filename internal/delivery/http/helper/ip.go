package helper

import (
	"net"
	"net/http"
	"strings"
)

// ExtractIP extracts the client IP from the request headers
func ExtractIP(r *http.Request) net.IP {
	// Cloudflare
	if cf := r.Header.Get("CF-Connecting-IP"); cf != "" {
		if ip := net.ParseIP(strings.TrimSpace(cf)); ip != nil {
			return ip
		}
	}

	// Akamai / proxies
	if tci := r.Header.Get("True-Client-IP"); tci != "" {
		if ip := net.ParseIP(strings.TrimSpace(tci)); ip != nil {
			return ip
		}
	}

	// Standard proxy header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if ip := net.ParseIP(trimmed); ip != nil {
				// Skip private IPs in X-Forwarded-For chain
				if !isPrivateIP(ip) {
					return ip
				}
			}
		}
	}

	// Nginx / Traefik
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		if ip := net.ParseIP(strings.TrimSpace(xrip)); ip != nil {
			return ip
		}
	}

	// Fallback to remote address
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		if ip := net.ParseIP(host); ip != nil {
			return ip
		}
	}

	return nil
}

func isPrivateIP(ip net.IP) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"::1/128",
		"fc00::/7",
	}

	for _, cidr := range privateRanges {
		_, subnet, _ := net.ParseCIDR(cidr)
		if subnet.Contains(ip) {
			return true
		}
	}
	return false
}
