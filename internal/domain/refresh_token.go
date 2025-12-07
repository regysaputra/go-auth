package domain

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a remember token.
type RefreshToken struct {
	ID           uuid.UUID  `json:"id"`
	UserID       int64      `json:"user_id"`
	TokenHash    string     `json:"token_hash"`
	ParentID     *uuid.UUID `json:"parent_id"`
	ReplacedBy   *uuid.UUID `json:"replaced_by"`
	RevokedAt    *time.Time `json:"revoked_at"`
	UserAgent    string     `json:"user_agent"`
	IpAddress    string     `json:"ip_address"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    time.Time  `json:"expires_at"`
	Country      *string    `json:"country"`
	Region       *string    `json:"region"`
	City         *string    `json:"city"`
	Latitude     *float64   `json:"latitude"`
	Longitude    *float64   `json:"longitude"`
	ASN          *uint      `json:"asn"`
	ASNOrg       *string    `json:"asn_org"`
	IsVPN        bool       `json:"is_vpn"`
	IsProxy      bool       `json:"is_proxy"`
	IsTor        bool       `json:"is_tor"`
	IsDatacenter bool       `json:"is_datacenter"`
}
