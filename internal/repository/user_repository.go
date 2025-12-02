package repository

import (
	"database/sql"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
)

type SQLUserRepository struct {
	db *sql.DB
}

func NewSQLUserRepository(DB *sql.DB) UserRepository {
	return &SQLUserRepository{
		db: DB,
	}
}

func (ur *SQLUserRepository) FindById(id string, user *domain.User) *utils.ReturnStatus {
	row := ur.db.QueryRow("SELECT * FROM users WHERE id = $1", id)
	err := row.Scan(&user.Id, &user.Username, &user.Password, &user.Email, &user.Role, &user.EnableTOTP, &user.SecretTOTP)

	if err != nil {
		return utils.ErrIfExists(utils.ErrCodeInternal, err)
	}

	return nil
}

func (ur *SQLUserRepository) FindByEmail(email string, user *domain.User) *utils.ReturnStatus {
	row := ur.db.QueryRow("SELECT * FROM users WHERE email = $1", email)
	err := row.Scan(&user.Id, &user.Username, &user.Password, &user.Email, &user.Role, &user.EnableTOTP, &user.SecretTOTP)
	if err != nil {
		return utils.ErrIfExists(utils.ErrCodeInternal, err)
	}

	return nil
}

func (ur *SQLUserRepository) AddTimestamp(id string, cid string) *utils.ReturnStatus {
	_, err := ur.db.Exec(`
		INSERT INTO usersLoginSession (id, cid)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE 
		SET cid = EXCLUDED.cid
	`, id, cid)

	return utils.ErrIfExists(utils.ErrCodeDatabaseError, err)
}
