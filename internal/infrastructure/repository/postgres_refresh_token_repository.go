package repository

import (
	"auth/internal/domain"
	"auth/internal/usecase"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRefreshTokenRepository represents the Postgres remember token repository object
type PostgresRefreshTokenRepository struct {
	db              *pgxpool.Pool
	refreshTokenTTL time.Duration
}

// NewPostgresRefreshTokenRepository creates a new Postgres remember token repository object
func NewPostgresRefreshTokenRepository(db *pgxpool.Pool, refreshTokenTTL time.Duration) usecase.RefreshTokenRepository {
	return &PostgresRefreshTokenRepository{db: db, refreshTokenTTL: refreshTokenTTL}
}

// Save the refresh token to the database
func (r *PostgresRefreshTokenRepository) Save(ctx context.Context, refreshToken domain.RefreshToken) (uuid.UUID, error) {
	expiresAt := time.Now().Add(r.refreshTokenTTL)

	sql := `
			INSERT INTO refresh_tokens (
				user_id,
				token_hash,
				parent_id,
				replaced_by,
				revoked_at,
				user_agent,
				ip_address,
				created_at,
				expires_at,
			    country,
			    region,
			    city,
			    latitude,
			    longitude,
			    asn,
			    asn_org,
			    is_vpn,
			    is_proxy,
			    is_tor,
			    is_datacenter
			) VALUES (
					$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
			) RETURNING id
	`
	row := r.db.QueryRow(
		ctx,
		sql,
		refreshToken.UserID,
		refreshToken.TokenHash,
		refreshToken.ParentID,
		refreshToken.ReplacedBy,
		refreshToken.RevokedAt,
		refreshToken.UserAgent,
		refreshToken.IpAddress,
		refreshToken.CreatedAt,
		expiresAt,
		refreshToken.Country,
		refreshToken.Region,
		refreshToken.City,
		refreshToken.Latitude,
		refreshToken.Longitude,
		refreshToken.ASN,
		refreshToken.ASNOrg,
		refreshToken.IsVPN,
		refreshToken.IsProxy,
		refreshToken.IsTor,
		refreshToken.IsDatacenter,
	)
	var id uuid.UUID

	err := row.Scan(&id)

	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

// FindByToken finds the refresh token by token hash
func (r *PostgresRefreshTokenRepository) FindByToken(ctx context.Context, hashToken string) (*domain.RefreshToken, error) {
	sql := "SELECT * FROM refresh_tokens WHERE token_hash = $1"

	row := r.db.QueryRow(ctx, sql, hashToken)
	var refreshToken domain.RefreshToken
	err := row.Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.TokenHash,
		&refreshToken.ParentID,
		&refreshToken.ReplacedBy,
		&refreshToken.RevokedAt,
		&refreshToken.UserAgent,
		&refreshToken.IpAddress,
		&refreshToken.CreatedAt,
		&refreshToken.ExpiresAt,
		&refreshToken.Country,
		&refreshToken.Region,
		&refreshToken.City,
		&refreshToken.Latitude,
		&refreshToken.Longitude,
		&refreshToken.ASN,
		&refreshToken.ASNOrg,
		&refreshToken.IsVPN,
		&refreshToken.IsProxy,
		&refreshToken.IsTor,
		&refreshToken.IsDatacenter,
	)

	if err != nil {
		return nil, err
	}

	return &refreshToken, nil
}

// Revoke invalidate refresh token before it would normally expire
func (r *PostgresRefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	sql := "UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1"
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

// Replace oldID with newID
func (r *PostgresRefreshTokenRepository) Replace(ctx context.Context, oldID, newID uuid.UUID) error {
	sql := "UPDATE refresh_tokens SET replaced_by = $1 WHERE id = $2"
	_, err := r.db.Exec(ctx, sql, newID, oldID)
	return err
}

// RevokeChain revokes all children of a given token and the token itself
func (r *PostgresRefreshTokenRepository) RevokeChain(ctx context.Context, id uuid.UUID) error {
	sql := `
		WITH RECURSIVE chain AS (
			SELECT id, parent_id
            FROM refresh_tokens
			WHERE id = $1
			
			UNION ALL
			
			SELECT rt.id, rt.parent_id
			FROM refresh_tokens rt
			INNER JOIN chain c ON rt.parent_id = c.id
		)
		UPDATE refresh_tokens SET revoked_at = NOW() WHERE id IN (SELECT id FROM chain)
	`

	_, err := r.db.Exec(ctx, sql, id)
	return err
}
