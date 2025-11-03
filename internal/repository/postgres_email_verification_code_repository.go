package repository

import (
	"auth/internal/domain"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresEmailVerificationCodeRepository represents the Postgres email verification code repository object
type PostgresEmailVerificationCodeRepository struct {
	db *pgxpool.Pool
}

// NewPostgresEmailVerificationCodeRepository creates a new Postgres email verification code repository object
func NewPostgresEmailVerificationCodeRepository(db *pgxpool.Pool) *PostgresEmailVerificationCodeRepository {
	return &PostgresEmailVerificationCodeRepository{db: db}
}

// GenerateCode generates a random code with the given length
func (r *PostgresEmailVerificationCodeRepository) GenerateCode(length int) (string, error) {
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

// Save saves the email verification code to the database
func (r *PostgresEmailVerificationCodeRepository) Save(ctx context.Context, email string, codeHash string, duration time.Duration) error {
	sql := "INSERT INTO email_verification_codes (email, code_hash, expires_at) VALUES ($1, $2, $3)"

	_, err := r.db.Exec(ctx, sql, email, codeHash, time.Now().Add(duration))

	return err
}

// Hash hashes the given code
func (r *PostgresEmailVerificationCodeRepository) Hash(code string) string {
	hash := sha256.Sum256([]byte(code))

	return fmt.Sprintf("%x", hash)
}

// FindByEmail finds the email verification code by email
func (r *PostgresEmailVerificationCodeRepository) FindByEmail(ctx context.Context, email string) (*domain.EmailVerificationCode, error) {
	sql := "SELECT * FROM email_verification_codes WHERE email = $1 AND expires_at > NOW()"
	row := r.db.QueryRow(ctx, sql, email)
	var ev domain.EmailVerificationCode
	err := row.Scan(&ev.ID, &ev.Email, &ev.CodeHash, &ev.ExpiresAt)

	if err != nil {
		return nil, err
	}

	return &ev, nil
}

// FindByCode finds the email verification code by code hash
func (r *PostgresEmailVerificationCodeRepository) FindByCode(ctx context.Context, codeHash string) (*domain.EmailVerificationCode, error) {
	sql := "SELECT * FROM email_verification_codes WHERE code_hash = $1 AND expires_at > NOW()"
	row := r.db.QueryRow(ctx, sql, codeHash)

	var ev domain.EmailVerificationCode
	err := row.Scan(&ev.ID, &ev.Email, &ev.CodeHash, &ev.ExpiresAt)

	if err != nil {
		return nil, err
	}

	return &ev, nil
}

// Delete deletes the email verification code by ID
func (r *PostgresEmailVerificationCodeRepository) Delete(ctx context.Context, id int64) error {
	sql := "DELETE FROM email_verification_codes WHERE id = $1"
	_, err := r.db.Exec(ctx, sql, id)
	return err
}
