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

type TestHandler struct {
	logger     *slog.Logger
	geoService *service.GeoIPService
}

func NewTestHandler(logger *slog.Logger, geoService *service.GeoIPService) *TestHandler {
	return &TestHandler{logger: logger, geoService: geoService}
}

func (h *TestHandler) GetGeoLocation(w http.ResponseWriter, r *http.Request) {
	ipAddress := helper.ExtractIP(r)

	var newGeo domain.GeoInfo
	if h.geoService != nil && ipAddress != nil {
		//ip := net.ParseIP(ipAddress.String())
		if net.ParseIP(ipAddress.String()) != nil {
			geoLookup, err := h.geoService.Lookup(ipAddress)

			if err == nil {
				newGeo = geoLookup
			}
		}
	}

	//fmt.Println("IP :", newGeo)
	//fmt.Println("Country :", newGeo)
	//fmt.Println("Region :", newGeo)
	//fmt.Println("City :", newGeo)
	//fmt.Println("Latitude :", newGeo)
	//fmt.Println("Longitude :", newGeo)
	//fmt.Println("ASN :", newGeo)
	//fmt.Println("ASNorg :", newGeo)
	//fmt.Println("ISP :", newGeo)
	//fmt.Println("isVPN :", newGeo)

	response.WriteSuccess(w, http.StatusOK, newGeo)
}
