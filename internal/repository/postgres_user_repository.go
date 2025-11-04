package repository

import (
	"auth/internal/domain"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresUserRepository represents the Postgres user repository object
type PostgresUserRepository struct {
	db *pgxpool.Pool
}

// NewPostgresUserRepository creates a new Postgres user repository object
func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// Save saves the user to the database
func (r *PostgresUserRepository) Save(ctx context.Context, user *domain.User) error {
	sql := "INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id"
	err := r.db.QueryRow(ctx, sql, user.Name, user.Email, user.Password).Scan(&user.ID)

	return err
}

// FindByEmail finds the user by email
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	fmt.Println("FindByEmail error :")
	sql := "SELECT * FROM users WHERE email = $1"
	row := r.db.QueryRow(ctx, sql, email)
	var user domain.User

	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Verified)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// FindByID finds the user by ID
func (r *PostgresUserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	sql := "SELECT * FROM users WHERE id = $1"
	row := r.db.QueryRow(ctx, sql, id)
	var user domain.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Verified)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// IsVerifiedUserExists checks if the user exists and is verified
func (r *PostgresUserRepository) IsVerifiedUserExists(ctx context.Context, email string) (bool, error) {
	sql := "SELECT EXISTS (SELECT 1 FROM users WHERE email = $1 AND verified = true)"
	var exist bool
	err := r.db.QueryRow(ctx, sql, email).Scan(&exist)

	fmt.Println("IsVerifiedUserExists Error :", err)

	if err != nil {
		return false, err
	}

	return exist, nil
}

// SetVerified sets the user as verified
func (r *PostgresUserRepository) SetVerified(ctx context.Context, userID int64) error {
	sql := "UPDATE users SET verified = TRUE WHERE id = $1"
	_, err := r.db.Exec(ctx, sql, userID)

	return err
}

// UpdatePassword updates the user's password'
func (r *PostgresUserRepository) UpdatePassword(ctx context.Context, userID int64, newPassword string) error {
	sql := "UPDATE users SET password = $1 WHERE id = $2"
	_, err := r.db.Exec(ctx, sql, newPassword, userID)
	return err
}
