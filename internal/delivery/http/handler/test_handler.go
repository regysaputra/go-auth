package handler

import (
	"auth/internal/delivery/http/helper"
	"auth/internal/delivery/http/response"
	"auth/internal/domain"
	"auth/internal/infrastructure/service"
	"log/slog"
	"net"
	"net/http"
)

// TestHandler represents the HTTP handler for test-related operations.
type TestHandler struct {
	logger     *slog.Logger
	geoService *service.GeoIPService
}

// NewTestHandler creates a new TestHandler instance.
func NewTestHandler(logger *slog.Logger, geoService *service.GeoIPService) *TestHandler {
	return &TestHandler{logger: logger, geoService: geoService}
}

// GetHeader retrieves the request headers.
func (h *TestHandler) GetHeader(w http.ResponseWriter, r *http.Request) {
	response.WriteSuccess(w, http.StatusOK, r.Header)
}

// GetGeoLocation retrieves the geolocation information for the client's IP address.
func (h *TestHandler) GetGeoLocation(w http.ResponseWriter, r *http.Request) {
	ipAddress := helper.ExtractIP(r)

	var newGeo domain.GeoInfo
	if h.geoService != nil && ipAddress != nil {
		if net.ParseIP(ipAddress.String()) != nil {
			geoLookup, err := h.geoService.Lookup(ipAddress)

			if err == nil {
				newGeo = geoLookup
			}
		}
	}

	response.WriteSuccess(w, http.StatusOK, newGeo)
}
