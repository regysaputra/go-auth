package domain

type GeoInfo struct {
	IP           string  `json:"ip"`
	Country      string  `json:"country"`
	Region       string  `json:"region"`
	City         string  `json:"city"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	ASN          uint    `json:"asn"`
	ASNOrg       string  `json:"asn_org"`
	ISP          string  `json:"isp"`
	IsVPN        bool    `json:"is_vpn"`
	IsProxy      bool    `json:"is_proxy"`
	IsTor        bool    `json:"is_tor"`
	IsDataCenter bool    `json:"is_data_center"`
}
