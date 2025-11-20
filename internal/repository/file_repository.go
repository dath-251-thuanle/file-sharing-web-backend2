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

	// LƯU Ý: Chúng ta phải giả định cột 'storage_name' không bắt buộc
	// vì nó không có trong lược đồ bạn cung cấp, nhưng cần cho việc xóa file vật lý.
	// Tạm thời, tôi bỏ qua nó trong INSERT.

	// Sử dụng file.Id và file.CreatedAt đã được Service thiết lập
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
			password, available_from, available_to, enable_totp, created_at
		FROM files
		WHERE id = $1
	`

	var file domain.File

	// Khai báo các biến sql.NullXxx cho các cột có thể NULL
	var ownerID sql.NullString
	var passwordHash sql.NullString

	row := r.db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&file.Id,
		&ownerID,       // user_id (NULLable)
		&file.FileName, // name
		&file.MimeType, // type
		&file.FileSize, // size
		&file.ShareToken,
		&passwordHash, // password (NULLable)
		&file.AvailableFrom,
		&file.AvailableTo,
		&file.EnableTOTP,
		&file.CreatedAt,
		&file.IsPublic,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err // Trả về sql.ErrNoRows nếu không tìm thấy
		}
		return nil, fmt.Errorf("failed to get file by ID: %w", err)
	}

	// Xử lý giá trị NULL sau khi Scan
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
	// SELECT * FROM files WHERE share_token = $1

	// return nil, sql.ErrNoRows // Mô phỏng

	query := `
		SELECT 
			id, user_id, name, type, size, share_token, 
			password, available_from, available_to, enable_totp, 
			created_at, is_public
		FROM files
		WHERE share_token = $1
	`

	var file domain.File

	// Khai báo các biến sql.NullXxx cho các cột có thể NULL
	var ownerID sql.NullString
	var passwordHash sql.NullString

	row := r.db.QueryRowContext(ctx, query, token)

	err := row.Scan(
		&file.Id,
		&ownerID,       // user_id (NULLable)
		&file.FileName, // name
		&file.MimeType, // type
		&file.FileSize, // size
		&file.ShareToken,
		&passwordHash, // password (NULLable)
		&file.AvailableFrom,
		&file.AvailableTo,
		&file.EnableTOTP,
		&file.CreatedAt,
		&file.IsPublic,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err // Trả về sql.ErrNoRows nếu không tìm thấy
		}
		return nil, err
	}

	// Xử lý giá trị NULL sau khi Scan
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
	// 1. Lệnh SQL DELETE file dựa trên ID và USER_ID
	query := `
        DELETE FROM files 
        WHERE id = $1 AND user_id = $2
    `

	// 2. Thực thi lệnh DELETE
	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to execute delete query for file ID %s: %w", id, err)
	}

	// 3. Kiểm tra số hàng bị xóa
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	// Nếu không có hàng nào bị ảnh hưởng (file không tồn tại hoặc userID không khớp)
	if rowsAffected == 0 {
		// Trả về lỗi, Service sẽ quyết định đây là 404 Not Found hay 403 Forbidden
		return sql.ErrNoRows
	}

	return nil
}

func (r *fileRepository) GetMyFiles(ctx context.Context, userID string, params domain.ListFileParams) ([]domain.File, error) {
	// LƯU Ý: Đây là logic truy vấn cơ bản, có thể cần thư viện sqlx để đơn giản hóa việc ánh xạ struct.

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

	// 2. Thêm điều kiện lọc Status
	if strings.ToLower(params.Status) != "all" {
		// Logic phức tạp hơn cần tính toán trạng thái (active, pending, expired) dựa trên thời gian.
		// Tạm thời bỏ qua Status Filter cho đến khi có trường Status trong DB hoặc có logic tính toán.
		// Nếu bạn có cột 'status' trong DB:
		// query += fmt.Sprintf(" AND status = $%d", argCounter)
		// args = append(args, params.Status)
		// argCounter++
	}

	// 3. Thêm sắp xếp
	// Chỉ cho phép sắp xếp theo cột hợp lệ để tránh SQL Injection
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
		// LƯU Ý: Phải sử dụng sql.NullTime cho các trường TIMESTAMPTZ NULLABLE
		// Phải sử dụng sql.NullString cho các trường TEXT/VARCHAR NULLABLE như password_hash, updated_at

		// Giả định các trường trong File struct đã được ánh xạ đúng
		err := rows.Scan(
			&f.Id, &f.OwnerId, &f.FileName, &f.MimeType, &f.FileSize, &f.ShareToken,
			&f.AvailableFrom, &f.AvailableTo, &f.EnableTOTP, &f.CreatedAt,
			&f.IsPublic,
			// ... Nếu có thêm cột cần scan
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file row: %w", err)
		}
		files = append(files, f)
	}

	return files, nil
}
