package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
)

type authRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepository{db: db}
}

func (ur *authRepository) Create(user *domain.User) (*domain.User, *utils.ReturnStatus) {
	row := ur.db.QueryRow("INSERT INTO users (id, username, password, email, role, enableTOTP, secretTOTP) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id", user.Id, user.Username, user.Password, user.Email, user.Role, user.EnableTOTP, user.SecretTOTP)
	err := row.Scan(&user.Id)
	fmt.Println("Created user with ID:", user.Id)
	if err != nil {
		return nil, utils.ResponseMsg(utils.ErrCodeDatabaseError, fmt.Sprintf("failed to create user: %w", err))
	}

	return user, nil
}

func (r *authRepository) BlacklistToken(token string, expiredAt time.Time) *utils.ReturnStatus {
	_, err := r.db.Exec(
		"INSERT INTO jwt_blacklist (token, expired_at) VALUES ($1, $2)",
		token, expiredAt,
	)

	return utils.ErrIfExists(utils.ErrCodeDatabaseError, err)
}

func (r *authRepository) IsTokenBlacklisted(token string) (bool, *utils.ReturnStatus) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM jwt_blacklist WHERE token = $1)",
		token,
	).Scan(&exists)

	return exists, utils.ErrIfExists(utils.ErrCodeDatabaseError, err)
}

func (r *authRepository) SaveSecret(userID string, secret string) *utils.ReturnStatus {
	_, err := r.db.Exec(`
		UPDATE users 
		SET secrettotp = $1 
		WHERE id = $2
	`, secret, userID)
	return utils.ErrIfExists(utils.ErrCodeDatabaseError, err)
}

func (r *authRepository) GetSecret(userID string) (string, *utils.ReturnStatus) {
	var secret string
	err := r.db.QueryRow(`
		SELECT secrettotp 
		FROM users 
		WHERE id = $1
	`, userID).Scan(&secret)

	if err != nil {
		return "", utils.ErrIfExists(utils.ErrCodeDatabaseError, err)
	}

	return secret, nil
}

func (r *authRepository) EnableTOTP(userID string) *utils.ReturnStatus {
	_, err := r.db.Exec(`UPDATE users SET "enabletotp" = TRUE WHERE id = $1`, userID)
	return utils.ErrIfExists(utils.ErrCodeDatabaseError, err)
}
