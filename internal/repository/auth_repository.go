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
	row := ur.db.QueryRow("INSERT INTO users (username, password, Email, Role, enableTOTP) VALUES ($1, $2, $3, $4, $5) RETURNING id", user.Username, user.Password, user.Email, user.Role, user.EnableTOTP)
	err := row.Scan(&user.Id)

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

func (r *authRepository) SaveSecret(userID string, secret string) error {
	_, err := r.db.Exec(`
		INSERT INTO user_totp(user_id, secret)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET secret = EXCLUDED.secret
	`, userID, secret)
	return err
}

func (r *authRepository) GetSecret(userID string) (string, error) {
	var secret string
	err := r.db.QueryRow(`SELECT secret FROM user_totp WHERE user_id = $1`, userID).Scan(&secret)
	if err != nil {
		return "", err
	}
	return secret, nil
}

func (r *authRepository) EnableTOTP(userID string) error {
	_, err := r.db.Exec(`UPDATE users SET "enableTOTP" = TRUE WHERE id = $1`, userID)
	return err
}