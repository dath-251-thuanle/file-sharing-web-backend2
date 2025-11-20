package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
)

type FileRepository interface {
	CreateFile(ctx context.Context, file *domain.File) (*domain.File, error)
	GetFileByID(ctx context.Context, id string) (*domain.File, error)
	GetFileByToken(ctx context.Context, token string) (*domain.File, error)
	DeleteFile(ctx context.Context, id string, userID string) error
	GetMyFiles(ctx context.Context, userID string, params domain.ListFileParams) ([]domain.File, error)
	FindAll(ctx context.Context) ([]domain.File, error)
}

type fileRepository struct {
	db *sql.DB
}

func NewFileRepository(db *sql.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) CreateFile(ctx context.Context, file *domain.File) (*domain.File, error) {
	// 1. Xử lý giá trị NULL cho cột UUID và Password
	var userID interface{}
	if file.OwnerId != nil {
		userID = *file.OwnerId
	} else {
		userID = nil // Anonymous Upload
	}

	// Cột 'password' trong DB cho phép NULL
	var passwordHash interface{}
	if file.PasswordHash != nil {
		passwordHash = *file.PasswordHash
	} else {
		passwordHash = nil
	}

	query := `
		INSERT INTO files (
			id, user_id, name, type, size, password, 
			available_from, available_to, enable_totp, 
			share_token, created_at, is_public
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		) RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query,
		file.Id,
		userID,             // $2: user_id (UUID hoặc NULL)
		file.FileName,      // $3: name
		file.MimeType,      // $4: type
		file.FileSize,      // $5: size
		passwordHash,       // $6: password (TEXT hoặc NULL)
		file.AvailableFrom, // $7: available_from
		file.AvailableTo,   // $8: available_to
		file.EnableTOTP,    // $9: enable_totp
		file.ShareToken,    // $10: share_token
		file.CreatedAt,     // $11: created_at,
		file.IsPublic,      // $12: is_public,
	).Scan(&file.Id, &file.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert file metadata: %w", err)
	}

	return file, nil
}

func (r *fileRepository) GetFileByID(ctx context.Context, id string) (*domain.File, error) {
	query := `
		SELECT 
			id, user_id, name, type, size, share_token, 
			password, available_from, available_to, enable_totp, created_at, is_public
		FROM files
		WHERE id = $1
	`

	var file domain.File

	var ownerID sql.NullString
	var passwordHash sql.NullString

	row := r.db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&file.Id,
		&ownerID,
		&file.FileName,
		&file.MimeType,
		&file.FileSize,
		&file.ShareToken,
		&passwordHash,
		&file.AvailableFrom,
		&file.AvailableTo,
		&file.EnableTOTP,
		&file.CreatedAt,
		&file.IsPublic,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get file by ID: %w", err)
	}

	if ownerID.Valid {
		file.OwnerId = &ownerID.String
	} else {
		file.OwnerId = nil
	}

	if passwordHash.Valid {
		file.PasswordHash = &passwordHash.String
		file.HasPassword = true
	} else {
		file.PasswordHash = nil
		file.HasPassword = false
	}

	return &file, nil
}

func (r *fileRepository) GetFileByToken(ctx context.Context, token string) (*domain.File, error) {
	query := `
		SELECT 
			id, user_id, name, type, size, share_token, 
			password, available_from, available_to, enable_totp, 
			created_at, is_public
		FROM files
		WHERE share_token = $1
	`

	var file domain.File
	var ownerID sql.NullString
	var passwordHash sql.NullString

	row := r.db.QueryRowContext(ctx, query, token)

	err := row.Scan(
		&file.Id,
		&ownerID,
		&file.FileName,
		&file.MimeType,
		&file.FileSize,
		&file.ShareToken,
		&passwordHash,
		&file.AvailableFrom,
		&file.AvailableTo,
		&file.EnableTOTP,
		&file.CreatedAt,
		&file.IsPublic,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}

	if ownerID.Valid {
		file.OwnerId = &ownerID.String
	} else {
		file.OwnerId = nil
	}

	if passwordHash.Valid {
		file.PasswordHash = &passwordHash.String
		file.HasPassword = true
	} else {
		file.PasswordHash = nil
		file.HasPassword = false
	}

	return &file, nil
}

func (r *fileRepository) DeleteFile(ctx context.Context, id string, userID string) error {
	// Giữ nguyên, vì nó chỉ DELETE metadata trong DB, việc xóa vật lý nằm ở Service
	query := `
        DELETE FROM files 
        WHERE id = $1 AND user_id = $2
    `

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to execute delete query for file ID %s: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *fileRepository) GetMyFiles(ctx context.Context, userID string, params domain.ListFileParams) ([]domain.File, error) {
	// 1. Khởi tạo truy vấn cơ bản
	baseQuery := `
		SELECT 
			id, user_id, name, type, size, share_token, 
			available_from, available_to, enable_totp, created_at, is_public
		FROM files
		WHERE user_id = $1 
	`
	args := []interface{}{userID}
	query := baseQuery
	argCounter := 2

	// 2. Thêm điều kiện lọc Status (giữ nguyên, vì không có cột status trong DB)
	if strings.ToLower(params.Status) != "all" {
		// ...
	}

	// 3. Thêm sắp xếp
	safeSortBy := "created_at"
	if params.SortBy == "fileName" {
		safeSortBy = "name"
	}
	safeOrder := "DESC"
	if strings.ToLower(params.Order) == "asc" {
		safeOrder = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", safeSortBy, safeOrder)

	// 4. Thêm phân trang (Pagination)
	offset := (params.Page - 1) * params.Limit
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, params.Limit, offset)

	// 5. Thực thi truy vấn
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query user files: %w", err)
	}
	defer rows.Close()

	var files []domain.File
	for rows.Next() {
		var f domain.File
		var ownerID sql.NullString // Cần để scan user_id

		err := rows.Scan(
			&f.Id, &ownerID, &f.FileName, &f.MimeType, &f.FileSize, &f.ShareToken,
			&f.AvailableFrom, &f.AvailableTo, &f.EnableTOTP, &f.CreatedAt,
			&f.IsPublic,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file row: %w", err)
		}

		// Gán giá trị sau khi scan
		if ownerID.Valid {
			f.OwnerId = &ownerID.String
		}

		files = append(files, f)
	}

	return files, nil
}
func (r *fileRepository) FindAll(ctx context.Context) ([]domain.File, error) {
	query := `
        SELECT 
            id, user_id, name, type, size, share_token, 
            password, available_from, available_to, enable_totp, created_at, is_public
        FROM files
        ORDER BY created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all files: %w", err)
	}
	defer rows.Close()

	var files []domain.File
	for rows.Next() {
		var f domain.File
		var ownerID sql.NullString
		var passwordHash sql.NullString

		err := rows.Scan(
			&f.Id,
			&ownerID,
			&f.FileName,
			&f.MimeType,
			&f.FileSize,
			&f.ShareToken,
			&passwordHash, // Cần password để xác định HasPassword
			&f.AvailableFrom,
			&f.AvailableTo,
			&f.EnableTOTP,
			&f.CreatedAt,
			&f.IsPublic,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan file row in FindAll: %w", err)
		}

		// Gán giá trị sau khi scan
		if ownerID.Valid {
			f.OwnerId = &ownerID.String
		} else {
			f.OwnerId = nil
		}

		if passwordHash.Valid {
			f.HasPassword = true
			f.PasswordHash = &passwordHash.String
		} else {
			f.HasPassword = false
			f.PasswordHash = nil
		}

		files = append(files, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iteration in FindAll: %w", err)
	}

	return files, nil
}
