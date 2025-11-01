package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresLoginOTPRepository struct {
	db *pgxpool.Pool
}

func NewPostgresLoginOTPRepository(db *pgxpool.Pool) *PostgresLoginOTPRepository {
	return &PostgresLoginOTPRepository{
		db: db,
	}
}

func (r *PostgresLoginOTPRepository) Generate(length int) (string, error) {
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

func (r *PostgresLoginOTPRepository) Hash(code string) string {
	hash := sha256.Sum256([]byte(code))

	return fmt.Sprintf("%x", hash)
}

func (r *PostgresLoginOTPRepository) Save(ctx context.Context, email string, codeHash string, duration time.Duration) error {
	sql := "INSERT INTO login_otps (email, code_hash, expires_at) VALUES ($1, $2, $3)"
	_, err := r.db.Exec(ctx, sql, email, codeHash, time.Now().Add(duration))

	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresLoginOTPRepository) IsCodeExist(ctx context.Context, codeHash string) error {
	sql := "SELECT EXISTS (SELECT 1 FROM login_otps WHERE code_hash = $1)"
	var exist bool
	err := r.db.QueryRow(ctx, sql, codeHash).Scan(&exist)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresLoginOTPRepository) Delete(ctx context.Context, email string) error {
	sql := "DELETE FROM login_otps WHERE email = $1"
	_, err := r.db.Exec(ctx, sql, email)
	if err != nil {
		return err
	}
	return nil
}
