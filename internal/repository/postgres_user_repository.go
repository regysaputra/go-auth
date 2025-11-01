package repository

import (
	"auth/internal/domain"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Save(ctx context.Context, user *domain.User) error {
	sql := "INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id"
	err := r.db.QueryRow(ctx, sql, user.Name, user.Email, user.Password).Scan(&user.ID)

	return err
}

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

func (r *PostgresUserRepository) FindById(ctx context.Context, id int64) (*domain.User, error) {
	sql := "SELECT * FROM users WHERE id = $1"
	row := r.db.QueryRow(ctx, sql, id)
	var user domain.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Verified)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

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

func (r *PostgresUserRepository) SetVerified(ctx context.Context, userID int64) error {
	sql := "UPDATE users SET verified = TRUE WHERE id = $1"
	_, err := r.db.Exec(ctx, sql, userID)

	return err
}

func (r *PostgresUserRepository) UpdatePassword(ctx context.Context, userID int64, newPassword string) error {
	sql := "UPDATE users SET password = $1 WHERE id = $2"
	_, err := r.db.Exec(ctx, sql, newPassword, userID)
	return err
}
