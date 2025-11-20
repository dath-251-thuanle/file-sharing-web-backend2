package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
)

type authRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepository{db: db}
}

func (ur *authRepository) Create(user *domain.User) (*domain.User, error) {
	row := ur.db.QueryRow("INSERT INTO users (userID, username, password, email, role, enableTOTP, secretTOTP) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id", user.Id, user.Username, user.Password, user.Email, user.Role, user.EnableTOTP, user.SecretTOTP)
	err := row.Scan(&user.Id)
	fmt.Println("Created user with ID:", user.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (r *authRepository) BlacklistToken(token string, expiredAt time.Time) error {
	_, err := r.db.Exec(
		"INSERT INTO jwt_blacklist (token, expired_at) VALUES ($1, $2)",
		token, expiredAt,
	)
	return err
}

func (r *authRepository) IsTokenBlacklisted(token string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM jwt_blacklist WHERE token = $1)",
		token,
	).Scan(&exists)

	return exists, err
}
