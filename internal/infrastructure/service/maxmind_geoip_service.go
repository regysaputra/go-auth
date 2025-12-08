package service

import (
	"auth/internal/domain"
	"errors"
	"net"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

// GeoIPService represents the geoip service object
type GeoIPService struct {
	cityDB *geoip2.Reader
	asnDB  *geoip2.Reader
}

// NewGeoIPService creates a new geoip service object
func NewGeoIPService(cityDBPath, asnDBPath string) (*GeoIPService, error) {
	cityDB, err := geoip2.Open(cityDBPath)
	if err != nil {
		return nil, err
	}

	asnDB, err := geoip2.Open(asnDBPath)
	if err != nil {
		//cityDB.Close()
		return nil, err
	}

	return &GeoIPService{
		cityDB: cityDB,
		asnDB:  asnDB,
	}, nil
}

// MustNewGeoIPService creates a new geoip service object or panics on error
func MustNewGeoIPService(cityDBPath, asnDBPath string) *GeoIPService {
	service, err := NewGeoIPService(cityDBPath, asnDBPath)
	if err != nil {
		panic(err)
	}

	return service
}

// Close closes the geoip databases
func (s *GeoIPService) Close() error {
	if s.cityDB == nil || s.asnDB == nil {
		return errors.New("geoip databases not loaded")
	}
	err := s.cityDB.Close()
	if err != nil {
		return err
	}
	err = s.asnDB.Close()
	if err != nil {
		return err
	}
	return nil
}

// Lookup performs a full geolocation query:
// - Country / Region / City
// - Latitude / Longitude
// - ASN (ISP & network number)
//
// (Correct MaxMind API usage — no nil deref, no extra args)
func (s *GeoIPService) Lookup(ip net.IP) (domain.GeoInfo, error) {
	// Validate input
	if ip == nil {
		return domain.GeoInfo{}, errors.New("invalid IP address")
	}

	// Initialize with IP
	info := domain.GeoInfo{IP: ip.String()}

	// --------------------
	// City database lookup
	// --------------------
	city, err := s.cityDB.City(ip)
	if err != nil {
		return domain.GeoInfo{}, err
	}

	// Country
	info.Country = city.Country.IsoCode

	// Region (first subdivision)
	if len(city.Subdivisions) > 0 {
		info.Region = city.Subdivisions[0].IsoCode
	}

	// City name (english version)
	if cityName, ok := city.City.Names["en"]; ok {
		info.City = cityName
	}

	// Coordinate
	info.Latitude = city.Location.Latitude
	info.Longitude = city.Location.Longitude

	// --------------------
	// ASN database lookup
	// --------------------
	asn, err := s.asnDB.ASN(ip)
	if err != nil {
		// ASN DB is optional, not fatal
		asn = nil
	}

	if asn != nil {
		info.ASN = asn.AutonomousSystemNumber
		info.ASNOrg = asn.AutonomousSystemOrganization
		info.ISP = asn.AutonomousSystemOrganization
	}

	// --------------------
	// Detection flags
	// --------------------
	info.IsDataCenter = s.isDataCenter(asn)
	info.IsVPN = s.isVPN(asn)
	info.IsProxy = s.isProxy(asn)
	info.IsTor = s.isTor(ip)

	return info, nil
}

// IsVPN ---------------------------
// VPN DETECTION
// ---------------------------
// Simple logic: IP belongs to known hosting / proxy ASN ranges.
// (You can expand this list later.)
func (s *GeoIPService) IsVPN(ip net.IP) bool {
	asn, err := s.asnDB.ASN(ip)
	if err != nil || asn == nil {
		return false
	}

	// Known data center ASNs (examples)
	knownVPNASNs := map[uint]bool{
		14061: true, // DigitalOcean
		16509: true, // AWS
		15169: true, // Google Cloud
		14618: true, // Amazon
		13335: true, // Cloudflare
		8075:  true, // Microsoft
	}

	_, isDataCenter := knownVPNASNs[asn.AutonomousSystemNumber]
	return isDataCenter
}

func (s *GeoIPService) isProxy(asn *geoip2.ASN) bool {
	if asn == nil {
		return false
	}

	org := strings.ToLower(asn.AutonomousSystemOrganization)
	proxyKeywords := []string{
		"proxy", "anonymizer", "hide", "shield",
	}

	for _, keyword := range proxyKeywords {
		if strings.Contains(org, keyword) {
			return true
		}
	}

	// If it's a data center but not residential, it could be a proxy
	return s.isDataCenter(asn)
}

// isDataCenter check if IP belongs to a known data center center/cloud provider
func (s *GeoIPService) isDataCenter(asn *geoip2.ASN) bool {
	if asn == nil {
		return false
	}

	// Known data center ASN
	dataCenterASN := map[uint]bool{
		// Cloud Providers
		14061:  true, // DigitalOcean
		16509:  true, // Amazon AWS
		15169:  true, // Google Cloud
		8075:   true, // Microsoft Azure
		13335:  true, // Cloudflare
		20473:  true, // Vultr
		24940:  true, // Hetzner
		14618:  true, // Amazon
		32934:  true, // Facebook
		396982: true, // Google Cloud

		// Other Hosting
		26496: true, // GoDaddy
		46606: true, // Unified Layer (Bluehost)
		63949: true, // Linode
		19318: true, // Interserver
		36351: true, // SoftLayer (IBM Cloud)
	}

	if dataCenterASN[asn.AutonomousSystemNumber] {
		return true
	}

	// Check for organizational name
	org := strings.ToLower(asn.AutonomousSystemOrganization)
	dataCenterKeywords := []string{
		"cloud", "hosting", "datacenter", "data center",
		"server", "colocation", "vps", "virtual",
	}

	for _, keyword := range dataCenterKeywords {
		if strings.Contains(org, keyword) {
			return true
		}
	}

	return false
}

// isVPN checks if IP belongs to a known VPN provider
func (s *GeoIPService) isVPN(asn *geoip2.ASN) bool {
	if asn == nil {
		return false
	}

	// Known VPN provider ASN
	vpnASN := map[uint]bool{
		// Major VPN Providers
		62041:  true, // NordVPN
		396998: true, // ExpressVPN
		201814: true, // Surfshark
		51396:  true, // Private Internet Access
		202425: true, // IP Vanish
		209103: true, // ProtonVPN
		211153: true, // CyberGhost

		// VPN Infrastructure
		44477: true, // Stark Industries (VPN)
		40676: true, // Psychz Networks (VPN hosting)
	}

	if vpnASN[asn.AutonomousSystemNumber] {
		return true
	}

	// Check by organization name
	org := strings.ToLower(asn.AutonomousSystemOrganization)
	vpnKeywords := []string{
		"vpn", "virtual private network", "nordvpn", "expressvpn",
		"surfshark", "protonvpn", "privateinternetaccess", "cyberghost",
		"tunnelbear", "vyprvpn", "ipvanish", "purevpn", "hidemyass",
	}

	for _, keyword := range vpnKeywords {
		if strings.Contains(org, keyword) {
			return true
		}
	}

	return false
}

// isTor checks if IP is a Tor exit node
// Note: This is a basic implementation. For production, use Tor's official exit node list
func (s *GeoIPService) isTor(ip net.IP) bool {
	// In production, you should:
	// 1. Download the Tor exit node list from: https://check.torproject.org/exit-addresses
	// 2. Parse and store it in memory/Redis
	// 3. Check if IP exists in that list

	if ip == nil {
		return false
	}
	// For now, return false (placeholder)
	// You can implement this by maintaining a list of known Tor exit nodes
	return false
}
