package repository

import (
	"auth/internal/domain"
	"auth/internal/usecase"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresLoginOTPRepository represents the postgres login otp repository object
type PostgresLoginOTPRepository struct {
	db          *pgxpool.Pool
	loginOTPTTL time.Duration
}

// NewPostgresLoginOTPRepository creates a new postgres login otp repository object
func NewPostgresLoginOTPRepository(db *pgxpool.Pool, loginOTPTTL time.Duration) usecase.LoginOTPRepository {
	return &PostgresLoginOTPRepository{
		db:          db,
		loginOTPTTL: loginOTPTTL,
	}
}

// GenerateCode generates a random string of length
func (r *PostgresLoginOTPRepository) GenerateCode(length int) (string, error) {
	var numbers = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	b := make([]byte, length)
	n, err := io.ReadAtLeast(rand.Reader, b, length)

	if n != length {
		return "", err
	}

	for i := 0; i < len(b); i++ {
		b[i] = numbers[int(b[i])%len(numbers)]
	}

	return string(b), nil
}

// HashCode hashes the code
func (r *PostgresLoginOTPRepository) HashCode(code string) string {
	hash := sha256.Sum256([]byte(code))

	return fmt.Sprintf("%x", hash)
}

// Save saves the code
func (r *PostgresLoginOTPRepository) Save(ctx context.Context, email string, codeHash string) error {
	expiresAt := time.Now().Add(r.loginOTPTTL)
	sql := "INSERT INTO login_otps (email, code_hash, expires_at) VALUES ($1, $2, $3)"
	_, err := r.db.Exec(ctx, sql, email, codeHash, expiresAt)

	if err != nil {
		return err
	}

	return nil
}

// IsCodeExist checks if the code exists
func (r *PostgresLoginOTPRepository) IsCodeExist(ctx context.Context, codeHash string) error {
	sql := "SELECT EXISTS (SELECT 1 FROM login_otps WHERE code_hash = $1)"
	var exist bool
	err := r.db.QueryRow(ctx, sql, codeHash).Scan(&exist)
	if err != nil {
		return err
	}

	return nil
}

// FindByCode find if code exists
func (r *PostgresLoginOTPRepository) FindByCode(ctx context.Context, codeHash string) (*domain.LoginOTP, error) {
	sql := "SELECT * FROM login_otps WHERE code_hash = $1"
	var loginOTP domain.LoginOTP
	err := r.db.QueryRow(ctx, sql, codeHash).Scan(&loginOTP.ID, &loginOTP.Email, &loginOTP.CodeHash, &loginOTP.ExpiresAt, &loginOTP.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &loginOTP, nil
}

// Delete deletes the code
func (r *PostgresLoginOTPRepository) Delete(ctx context.Context, email string) error {
	sql := "DELETE FROM login_otps WHERE email = $1"
	_, err := r.db.Exec(ctx, sql, email)
	if err != nil {
		return err
	}
	return nil
}
