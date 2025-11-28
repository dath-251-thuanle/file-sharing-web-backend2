package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
)

type FileRepository interface {
	CreateFile(ctx context.Context, file *domain.File) (*domain.File, *utils.ReturnStatus)
	GetFileByID(ctx context.Context, id string) (*domain.File, *utils.ReturnStatus)
	GetFileByToken(ctx context.Context, token string) (*domain.File, *utils.ReturnStatus)
	DeleteFile(ctx context.Context, id string, userID string) *utils.ReturnStatus
	GetMyFiles(ctx context.Context, userID string, params domain.ListFileParams) ([]domain.File, *utils.ReturnStatus)
	GetTotalUserFiles(ctx context.Context, userID string) (int, *utils.ReturnStatus)
	GetFileSummary(ctx context.Context, userID string) (*domain.FileSummary, *utils.ReturnStatus)
	FindAll(ctx context.Context) ([]domain.File, *utils.ReturnStatus)
	RegisterDownload(ctx context.Context, fileID string, userID string) *utils.ReturnStatus
	GetFileDownloadHistory(ctx context.Context, fileID string) (*domain.FileDownloadHistory, *utils.ReturnStatus)
	GetFileStats(ctx context.Context, fileID string) (*domain.FileStat, *utils.ReturnStatus)
}

type fileRepository struct {
	db *sql.DB
}

func NewFileRepository(db *sql.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) CreateFile(ctx context.Context, file *domain.File) (*domain.File, *utils.ReturnStatus) {
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
		return nil, utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
	}

	if _, err := r.db.Exec(`INSERT INTO filestat (file_id) VALUES ($1)`, file.Id); err != nil {
		return nil, utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
	}

	return file, nil
}

func (r *fileRepository) GetFileByID(ctx context.Context, id string) (*domain.File, *utils.ReturnStatus) {
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
			return nil, utils.Response(utils.ErrCodeFileNotFound)
		}
		return nil, utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
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

func (r *fileRepository) GetFileByToken(ctx context.Context, token string) (*domain.File, *utils.ReturnStatus) {
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
			return nil, utils.Response(utils.ErrCodeFileNotFound)
		}
		return nil, utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
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

func (r *fileRepository) DeleteFile(ctx context.Context, id string, userID string) *utils.ReturnStatus {
	query := `
        DELETE FROM files 
        WHERE id = $1 AND user_id = $2
    `

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
	}

	if rowsAffected == 0 {
		return utils.Response(utils.ErrCodeFileNotFound)
	}

	return nil
}

func (r *fileRepository) GetMyFiles(ctx context.Context, userID string, params domain.ListFileParams) ([]domain.File, *utils.ReturnStatus) {
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
		// LƯU Ý: Đây là logic lọc trạng thái (Status) trong truy vấn SQL chính.
		status := strings.ToLower(params.Status)

		// Tăng bộ đếm tham số
		argCounter++

		// Bổ sung điều kiện WHERE dựa trên Status
		if status == "active" {
			query += fmt.Sprintf(" AND available_from <= NOW() AND available_to > NOW()")
		} else if status == "pending" {
			query += fmt.Sprintf(" AND available_from > NOW()")
		} else if status == "expired" {
			query += fmt.Sprintf(" AND available_to <= NOW()")
		}
		// Nếu status không khớp, truy vấn sẽ không thay đổi, chỉ lọc user_id.
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
	query += fmt.Sprintf(" LIMIT $2 OFFSET $3")
	args = append(args, int64(params.Limit), int64(offset))

	// 5. Thực thi truy vấn
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
	}
	defer rows.Close()

	now := time.Now()

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
			return nil, utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
		}

		// Gán giá trị sau khi scan
		if ownerID.Valid {
			f.OwnerId = &ownerID.String
		}

		f.Status = "active"

		if now.Before(f.AvailableFrom) {
			f.Status = "pending"
		} else if now.After(f.AvailableTo) {
			f.Status = "expired"
		}

		files = append(files, f)
	}

	return files, nil
}
func (r *fileRepository) GetTotalUserFiles(ctx context.Context, userID string) (int, *utils.ReturnStatus) {
	var total int

	query := `SELECT COUNT(id) FROM files WHERE user_id = $1`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&total)
	if err != nil {
		return 0, utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
	}

	return total, nil
}
func (r *fileRepository) GetFileSummary(ctx context.Context, userID string) (*domain.FileSummary, *utils.ReturnStatus) {
	summary := &domain.FileSummary{}

	activeQuery := `
        SELECT COUNT(id) FROM files 
        WHERE user_id = $1 
          AND available_from <= NOW() 
          AND available_to > NOW()
    `
	err := r.db.QueryRowContext(ctx, activeQuery, userID).Scan(&summary.ActiveFiles) // Chỉ truyền $1
	if err != nil {
		return nil, utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
	}

	pendingQuery := `
        SELECT COUNT(id) FROM files 
        WHERE user_id = $1 
          AND available_from > NOW()
    `
	err = r.db.QueryRowContext(ctx, pendingQuery, userID).Scan(&summary.PendingFiles) // Chỉ truyền $1
	if err != nil {
		return nil, utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
	}

	// 3. Tính Expired Files (Đã hết hiệu lực: NOW >= available_to)
	expiredQuery := `
        SELECT COUNT(id) FROM files 
        WHERE user_id = $1 
          AND available_to <= NOW()
    `
	err = r.db.QueryRowContext(ctx, expiredQuery, userID).Scan(&summary.ExpiredFiles) // Chỉ truyền $1
	if err != nil {
		return nil, utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
	}

	return summary, nil
}

func (r *fileRepository) FindAll(ctx context.Context) ([]domain.File, *utils.ReturnStatus) {
	query := `
        SELECT 
            id, user_id, name, type, size, share_token, 
            password, available_from, available_to, enable_totp, created_at, is_public
        FROM files
        ORDER BY created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
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
			return nil, utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
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
		return nil, utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
	}

	return files, nil
}

func (r *fileRepository) RegisterDownload(ctx context.Context, fileID string, userID string) *utils.ReturnStatus {
	_, err := r.db.ExecContext(ctx, `CALL proc_download($1, $2)`, fileID, userID)

	if err != nil {
		return utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
	}

	return nil
}

func (r *fileRepository) GetFileDownloadHistory(ctx context.Context, fileID string) (*domain.FileDownloadHistory, *utils.ReturnStatus) {
	file, err := r.GetFileByID(ctx, fileID)
	if err != nil {
		log.Println("File retrieval failure")
		return nil, err
	}

	history := domain.FileDownloadHistory{}
	history.FileId = file.Id
	history.FileName = file.FileName

	rows, derr := r.db.QueryContext(ctx, `SELECT download_id, user_id, time FROM download WHERE file_id = $1`, file.Id)
	if derr != nil {
		log.Println("Download retrieval failure")
		return nil, utils.ResponseMsg(utils.ErrCodeDatabaseError, derr.Error())
	}

	for rows.Next() {
		var time time.Time
		var d_id string
		var u_id string
		if err := rows.Scan(&d_id, &u_id, &time); err != nil {
			log.Println("Row scan failure")
			return nil, utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
		}

		history.History = append(history.History,
			domain.Download{
				DownloadId:        d_id,
				UserId:            &u_id,
				Downloader:        nil,
				DownloadedAt:      time,
				DownloadCompleted: true,
			})
	}

	return &history, nil
}

func (r *fileRepository) GetFileStats(ctx context.Context, fileID string) (*domain.FileStat, *utils.ReturnStatus) {
	query := `
        SELECT 
            f.id, 
            f.user_id, 
            f.name, 
            COALESCE(s.download_count, 0), 
            COALESCE(s.user_download_count, 0), 
            f.created_at, 
            MAX(d.time)
        FROM files f
        LEFT JOIN filestat s ON f.id = s.file_id
        LEFT JOIN download d ON f.id = d.file_id
        WHERE f.id = $1
        GROUP BY f.id, f.user_id, f.name, s.download_count, s.user_download_count, f.created_at
    `

	stat := domain.FileStat{}
	var ownerID sql.NullString
	var lastDownloadTime sql.NullTime
	row := r.db.QueryRowContext(ctx, query, fileID)

	err := row.Scan(
		&stat.FileId,
		&ownerID, // Hứng vào NullString
		&stat.FileName,
		&stat.TotalDownloadCount,
		&stat.UserDownloadCount,
		&stat.CreatedAt,
		&lastDownloadTime, // Hứng vào NullTime
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.Response(utils.ErrCodeFileNotFound)
		}
		fmt.Println("DB Error:", err)
		return nil, utils.ResponseMsg(utils.ErrCodeDatabaseError, err.Error())
	}

	if lastDownloadTime.Valid {
		stat.LastDownloadedAt = lastDownloadTime.Time
	}

	return &stat, nil
}
