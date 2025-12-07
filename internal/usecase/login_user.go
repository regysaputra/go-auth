package usecase

import (
	"auth/internal/domain"
	"context"
	"database/sql"
	"errors"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type LoginUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type LoginResult struct {
	LoginUserObj    *LoginUser
	AccessToken     string
	RefreshToken    string
	RefreshTokenExp time.Time
}

// LoginUserUseCase represents the login user use case object
type LoginUserUseCase struct {
	userRepository         UserRepository
	tokenRepository        TokenRepository
	refreshTokenRepository RefreshTokenRepository
	geoIPService           GeoIPService
	accessTokenTTL         time.Duration
	refreshTokenTTL        time.Duration
}

// NewLoginUserUseCase creates a new login user use case object
func NewLoginUserUseCase(
	userRepository UserRepository,
	tokenRepository TokenRepository,
	refreshTokenRepository RefreshTokenRepository,
	geoIPService GeoIPService,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *LoginUserUseCase {
	return &LoginUserUseCase{
		userRepository:         userRepository,
		tokenRepository:        tokenRepository,
		refreshTokenRepository: refreshTokenRepository,
		geoIPService:           geoIPService,
		accessTokenTTL:         accessTokenTTL,
		refreshTokenTTL:        refreshTokenTTL,
	}
}

// Execute authenticates a user by checking their credentials and then generates tokens for them
func (uc *LoginUserUseCase) Execute(ctx context.Context, email string, password string, rememberMe bool, userAgent string, ipAddress net.IP) (*LoginResult, error) {
	// Field validation
	validationErrors := NewValidationErrors()
	email = strings.TrimSpace(email)
	if email == "" {
		validationErrors.Add("email", "email field is required")
	} else if !strings.Contains(email, "@") {
		validationErrors.Add("email", "email must be a valid email address")
	}

	if password == "" {
		validationErrors.Add("password", "password field is required")
	} else if len(password) < 8 {
		validationErrors.Add("password", "password must be at least 8 characters")
	}

	if validationErrors.HasErrors() {
		return nil, validationErrors
	}

	// find a user by email
	user, err := uc.userRepository.FindByEmail(ctx, email)
	if err != nil {
		// Don't return user not found error to prevent email enumeration attack
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	// compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return uc.GenerateLoginToken(ctx, user, rememberMe, userAgent, ipAddress)
}

// GenerateLoginToken Creates a new JWT and optionally a remembered me token for a given user ID
// This method is separate from Execute so it can be called directly after other authentication flows, like email verification.
func (uc *LoginUserUseCase) GenerateLoginToken(ctx context.Context, user *domain.User, rememberMe bool, userAgent string, ipAddress net.IP) (*LoginResult, error) {
	// generate access token for authenticated user
	accessToken, err := uc.tokenRepository.GenerateJWTToken(user.ID, uc.accessTokenTTL)

	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.tokenRepository.GenerateOpaqueToken()
	if err != nil {
		return nil, err
	}

	refreshTokenHash := uc.tokenRepository.HashToken(refreshToken)

	var geo domain.GeoInfo
	if uc.geoIPService != nil && ipAddress != nil {
		ip := net.ParseIP(ipAddress.String())
		if ip != nil {
			geoLookup, err := uc.geoIPService.Lookup(ip)
			if err == nil {
				geo = geoLookup
			}
		}
	}

	refreshTokenObj := domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		UserAgent: userAgent,
		IpAddress: ipAddress.String(),
		CreatedAt: time.Now(),
		Country:   &geo.Country,
		Region:    &geo.Region,
		City:      &geo.City,
		Latitude:  &geo.Latitude,
		Longitude: &geo.Longitude,
	}

	if rememberMe {
		refreshTokenObj.ExpiresAt = time.Now().Add(uc.refreshTokenTTL * 30)
	} else {
		refreshTokenObj.ExpiresAt = time.Now().Add(uc.refreshTokenTTL)
	}

	_, err = uc.refreshTokenRepository.Save(ctx, refreshTokenObj)
	if err != nil {
		return nil, err
	}

	loginResult := &LoginResult{
		LoginUserObj:    &LoginUser{Name: user.Name, Email: user.Email},
		AccessToken:     accessToken,
		RefreshToken:    refreshToken,
		RefreshTokenExp: refreshTokenObj.ExpiresAt,
	}

	return loginResult, nil
}
