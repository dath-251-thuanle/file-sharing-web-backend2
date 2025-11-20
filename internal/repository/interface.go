package repository

import (
	"time"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
)

type UserRepository interface {
	FindById(id string, user *domain.User) error
	FindByEmail(email string, user *domain.User) error
}

type AuthRepository interface {
	BlacklistToken(token string, expiredAt time.Time) error
	IsTokenBlacklisted(token string) (bool, error)
	Create(user *domain.User) (*domain.User, error)
	SaveSecret(userID string, secret string) error
	GetSecret(userID string) (string, error)
	EnableTOTP(userID string) error
}
