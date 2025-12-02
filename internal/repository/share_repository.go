package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
)

type SharedRepository interface {
	ShareFileWithUsers(ctx context.Context, fileID string, emails []string) *utils.ReturnStatus
	GetUsersSharedWith(ctx context.Context, fileID string) (*domain.Shared, *utils.ReturnStatus)
}

type sharedRepository struct {
	db *sql.DB
}

func NewSharedRepository(db *sql.DB) SharedRepository {
	return &sharedRepository{db: db}
}

func (r *sharedRepository) ShareFileWithUsers(ctx context.Context, fileID string, emails []string) *utils.ReturnStatus {
	if len(emails) == 0 {
		return nil
	}

	var emailstrings []string
	for _, e := range emails {
		emailstrings = append(emailstrings, fmt.Sprintf("'%s'", e))
	}

	userIDQuery := fmt.Sprintf(`SELECT id FROM users WHERE email IN (%s);`, strings.Join(emailstrings, ", "))

	userIDsRaw, err := r.db.QueryContext(ctx, userIDQuery)
	if err != nil {
		log.Println("Email retrieval failure")
		return utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
	}

	var queryValues []string
	for userIDsRaw.Next() {
		var userid_tmp string
		if err := userIDsRaw.Scan(&userid_tmp); err != nil {
			log.Println("Email scan failure")
			return utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
		}

		queryValues = append(queryValues, fmt.Sprintf("('%s', '%s')", userid_tmp, fileID))
	}

	if len(queryValues) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
		INSERT INTO shared (user_id, file_id) 
		VALUES %s 
		ON CONFLICT (user_id, file_id) DO NOTHING
	`, strings.Join(queryValues, ", "))

	if _, err := r.db.ExecContext(ctx, query); err != nil {
		log.Println("INSERT failure")
		return utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
	}

	return nil
}

func (r *sharedRepository) GetUsersSharedWith(ctx context.Context, fileID string) (*domain.Shared, *utils.ReturnStatus) {
	// SELECT * FROM shared_with WHERE file_id = $1

	query := `
		SELECT user_id FROM shared WHERE file_id = $1
	`

	share := domain.Shared{
		FileId:  fileID,
		UserIds: make([]string, 0, 10),
	}

	rows, err := r.db.QueryContext(ctx, query, fileID)
	if err != nil {
		log.Println(err)
		return nil, utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
	}

	for rows.Next() {
		var userid_tmp string

		if err := rows.Scan(&userid_tmp); err != nil {
			log.Println(err)
			return nil, utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
		}

		share.UserIds = append(share.UserIds, userid_tmp)
	}

	return &share, nil
}
