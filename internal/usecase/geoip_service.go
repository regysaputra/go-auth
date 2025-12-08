package usecase

import (
	"auth/internal/domain"
	"net"
)

// GeoIPService interface
type GeoIPService interface {
	Lookup(ip net.IP) (domain.GeoInfo, error)
	IsVPN(ip net.IP) bool
	Close() error
}
