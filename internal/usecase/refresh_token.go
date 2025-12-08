package usecase

import (
	"auth/internal/domain"
	"auth/internal/infrastructure/service"
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net"
	"time"
)

// RefreshTokenResult Hold the output of a successful token refresh
type RefreshTokenResult struct {
	AccessToken     string    `json:"access_token"`
	RefreshToken    string    `json:"refresh_token"`
	RefreshTokenExp time.Time `json:"refresh_token_exp"`
}

// RefreshTokenUseCase handle the logic for exchanging a remember token for a new JWT
type RefreshTokenUseCase struct {
	userRepository          UserRepository
	refreshTokenRepository  RefreshTokenRepository
	tokenRepository         TokenRepository
	geoService              *service.GeoIPService
	logger                  *slog.Logger
	accessTokenTTL          time.Duration
	riskRevokeThreshold     int
	riskInvalidateThreshold int
}

// NewRefreshTokenUseCase creates a new RefreshTokenUseCase object
func NewRefreshTokenUseCase(
	userRepository UserRepository,
	refreshTokenRepository RefreshTokenRepository,
	tokenRepository TokenRepository,
	geoService *service.GeoIPService,
	logger *slog.Logger,
	accessTokenTTL time.Duration,
	riskRevokeThreshold int,
	riskInvalidateThreshold int,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		userRepository:          userRepository,
		refreshTokenRepository:  refreshTokenRepository,
		tokenRepository:         tokenRepository,
		geoService:              geoService,
		logger:                  logger,
		accessTokenTTL:          accessTokenTTL,
		riskRevokeThreshold:     riskRevokeThreshold,
		riskInvalidateThreshold: riskInvalidateThreshold,
	}
}

// Execute validates a refresh token, performs secure token rotation and issues a new JWT
func (uc *RefreshTokenUseCase) Execute(
	ctx context.Context,
	refreshToken string,
	userAgent string,
	ipAddress net.IP,
) (*RefreshTokenResult, error) {
	if refreshToken == "" {
		return nil, ErrInvalidToken
	}

	// Find refresh token in database
	refreshTokenHash := uc.tokenRepository.HashToken(refreshToken)
	oldRefreshToken, err := uc.refreshTokenRepository.FindByToken(ctx, refreshTokenHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidToken
		}

		return nil, err
	}

	// Verify if refresh token still valid
	if oldRefreshToken.RevokedAt != nil {
		// Revoke entire chain
		err := uc.refreshTokenRepository.RevokeChain(ctx, oldRefreshToken.ID)

		if err != nil {
			return nil, err
		}

		return nil, ErrInvalidToken
	}
	if oldRefreshToken.ExpiresAt.Before(time.Now()) {
		return nil, ErrInvalidToken
	}

	// Lookup geo for current request
	var newGeo domain.GeoInfo
	if uc.geoService != nil && ipAddress != nil {
		ip := net.ParseIP(string(ipAddress))
		if ip != nil {
			geoLookup, err := uc.geoService.Lookup(ip)
			if err == nil {
				newGeo = geoLookup
			}
		}
	}

	// Calculate risk score
	risk, reasons := uc.calculateRisk(oldRefreshToken, &newGeo, userAgent)
	uc.logger.Info("Reasons", "reasons", reasons)

	// Act on risk
	if risk >= uc.riskRevokeThreshold {
		err = uc.refreshTokenRepository.RevokeChain(ctx, oldRefreshToken.ID)
		if err != nil {
			return nil, err
		}

		return nil, ErrSuspiciousActivity
	}

	if risk >= uc.riskInvalidateThreshold {
		err = uc.refreshTokenRepository.Revoke(ctx, oldRefreshToken.ID)
		if err != nil {
			return nil, err
		}

		return nil, ErrSuspiciousActivity
	}

	// Rotation: Issue a new token pair
	newRefreshToken, err := uc.tokenRepository.GenerateOpaqueToken()

	if err != nil {
		return nil, err
	}

	newRefreshTokenHash := uc.tokenRepository.HashToken(newRefreshToken)
	newAccessToken, err := uc.tokenRepository.GenerateJWTToken(oldRefreshToken.UserID, uc.accessTokenTTL)

	if err != nil {
		return nil, err
	}

	// Create new refresh token object
	newRefreshTokenObj := domain.RefreshToken{
		UserID:    oldRefreshToken.UserID,
		TokenHash: newRefreshTokenHash,
		ParentID:  &oldRefreshToken.ID,
		UserAgent: userAgent,
		IPAddress: ipAddress.String(),
		CreatedAt: time.Now(),
		ExpiresAt: oldRefreshToken.ExpiresAt,
	}

	// Populate geo fields if available (safe nil handling)
	if newGeo.Country != "" {
		newRefreshTokenObj.Country = &newGeo.Country
	}
	if newGeo.Region != "" {
		newRefreshTokenObj.Region = &newGeo.Region
	}
	if newGeo.City != "" {
		newRefreshTokenObj.City = &newGeo.City
	}
	if newGeo.Latitude != 0 {
		newRefreshTokenObj.Latitude = &newGeo.Latitude
	}
	if newGeo.Longitude != 0 {
		newRefreshTokenObj.Longitude = &newGeo.Longitude
	}
	if newGeo.Latitude != 0 {
		newRefreshTokenObj.Latitude = &newGeo.Latitude
	}
	// ASN/org flags (if your domain fields are pointer/ints)
	if newGeo.ASN != 0 {
		asn := newGeo.ASN
		newRefreshTokenObj.ASN = &asn
	}
	if newGeo.ASNOrg != "" {
		newRefreshTokenObj.ASNOrg = &newGeo.ASNOrg
	}
	newRefreshTokenObj.IsVPN = newGeo.IsVPN
	newRefreshTokenObj.IsProxy = newGeo.IsProxy
	newRefreshTokenObj.IsTor = newGeo.IsTor
	newRefreshTokenObj.IsDatacenter = newGeo.IsDataCenter

	// Insert new refresh token to database
	id, err := uc.refreshTokenRepository.Save(ctx, newRefreshTokenObj)

	if err != nil {
		return nil, err
	}

	// Link old -> new and revoke old
	err = uc.refreshTokenRepository.Replace(ctx, oldRefreshToken.ID, id)

	if err != nil {
		return nil, err
	}

	err = uc.refreshTokenRepository.Revoke(ctx, oldRefreshToken.ID)

	if err != nil {
		return nil, err
	}

	result := &RefreshTokenResult{
		AccessToken:     newAccessToken,
		RefreshToken:    newRefreshToken,
		RefreshTokenExp: oldRefreshToken.ExpiresAt,
	}

	return result, nil
}

// calculateRisk compute a simple risk score and reasons
func (uc *RefreshTokenUseCase) calculateRisk(oldRefreshToken *domain.RefreshToken, newGeo *domain.GeoInfo, userAgent string) (int, []string) {
	var score int
	var reasons []string

	// Impossible travel (strong)
	if oldRefreshToken.Longitude != nil && oldRefreshToken.Latitude != nil && newGeo != nil && newGeo.Longitude != 0 && newGeo.Latitude != 0 {
		prev := domain.GeoInfo{
			Country:   "",
			Region:    "",
			City:      "",
			Latitude:  *oldRefreshToken.Latitude,
			Longitude: *oldRefreshToken.Longitude,
		}

		prevTime := oldRefreshToken.CreatedAt
		currentTime := time.Now()

		impossible, detail := domain.ImpossibleTravel(prev, prevTime, *newGeo, currentTime, domain.DefaultConfig)

		if impossible {
			score += 50
			reasons = append(reasons, "impossible_travel:"+detail)
		}
	}

	// Country change (medium)
	if oldRefreshToken.Country != nil && newGeo != nil && newGeo.Country != "" && *oldRefreshToken.Country != newGeo.Country {
		score += 20
		reasons = append(reasons, "country_changed")
	}

	// 3) VPN/datacenter/proxy/tor (weak -> medium)
	if newGeo != nil {
		if newGeo.IsVPN {
			score += 15
			reasons = append(reasons, "vpn_detected")
		}
		if newGeo.IsDataCenter {
			score += 15
			reasons = append(reasons, "datacenter_ip")
		}
		if newGeo.IsProxy || newGeo.IsTor {
			score += 25
			reasons = append(reasons, "proxy_or_tor")
		}
	}

	// 4) ASN change (mild)
	if oldRefreshToken.ASN != nil && newGeo != nil && newGeo.ASN != 0 && *oldRefreshToken.ASN != newGeo.ASN {
		score += 10
		reasons = append(reasons, "asn_changed")
	}

	// 5) User-Agent change (light by itself; heavier when combined)
	if oldRefreshToken.UserAgent != "" && userAgent != "" && oldRefreshToken.UserAgent != userAgent {
		// base penalty low
		score += 10
		reasons = append(reasons, "user_agent_changed")
		// increase if combined with other signals:
		if (newGeo != nil && newGeo.IsVPN) || (newGeo != nil && newGeo.IsDataCenter) {
			score += 10
			reasons = append(reasons, "ua_plus_vpn_datacenter")
		}
	}

	return score, reasons
}
