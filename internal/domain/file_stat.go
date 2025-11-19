package domain

type FileStat struct {
	Id                 string `json:"id" db:"id"`
	FileId             string `json:"fileId" db:"file_id"`
	IsPublic           bool   `json:"isPublic" db:"is_public"`
	Status             string `json:"status" db:"status"` // pending | active | expired
	UserDownloadCount  int    `json:"userDownloadCount" db:"user_download_count"`
	TotalDownloadCount int    `json:"totalDownloadCount" db:"total_download_count"` // Bao gồm cả người dùng ẩn danh
}
