package usecase

import (
	"auth/internal/domain"
	"net"
)

type GeoIPService interface {
	Lookup(ip net.IP) (domain.GeoInfo, error)
	IsVPN(ip net.IP) bool
	Close() error
}
