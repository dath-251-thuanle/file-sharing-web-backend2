package domain

import "time"

type FileStatus string

const (
	FILE_PENDING FileStatus = "pending"
	FILE_ACTIVE  FileStatus = "active"
	FILE_EXPIRED FileStatus = "expired"
)

type File struct {
	Id            string     `json:"id" db:"id"`
	OwnerId       *string    `json:"ownerId" db:"user_id"`
	FileName      string     `json:"fileName" db:"name"`
	StorageName   string     `json:"-" db:"storage_name"`
	FileSize      int64      `json:"fileSize" db:"size"`
	MimeType      string     `json:"mimeType" db:"type"`
	ShareToken    string     `json:"shareToken" db:"share_token"`
	IsPublic      bool       `json:"isPublic" db:"is_public"`
	HasPassword   bool       `json:"hasPassword" db:"has_password"`
	PasswordHash  *string    `json:"-" db:"password"`
	EnableTOTP    bool       `json:"-" db:"enable_totp"`
	AvailableFrom time.Time  `json:"availableFrom" db:"available_from"`
	AvailableTo   time.Time  `json:"availableTo" db:"available_to"`
	ValidityDays  int        `json:"-" db:"validity_days"`
	Status        FileStatus `json:"status"`
	CreatedAt     time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt     *time.Time `json:"-" db:"updated_at"`
}

type Pagination struct {
	CurrentPage  int `json:"currentPage"`
	TotalPages   int `json:"totalPages"`
	TotalRecords int `json:"totalRecords"`
	Limit        int `json:"limit"`
}

type Downloader struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type Download struct {
	DownloadId        string      `json:"id"`
	UserId            *string     `json:"-"`
	Downloader        *Downloader `json:"downloader"`
	DownloadedAt      time.Time   `json:"downloadedAt"`
	DownloadCompleted bool        `json:"downloadCompleted"`
}

type FileDownloadHistory struct {
	FileId     string     `json:"fileId"`
	FileName   string     `json:"fileName"`
	History    []Download `json:"history"`
	Pagination Pagination `json:"pagination"`
}

type ListFileParams struct {
	Status string
	Page   int
	Limit  int
	SortBy string
	Order  string
}

type FileSummary struct {
	ActiveFiles  int `json:"activeFiles"`
	PendingFiles int `json:"pendingFiles"`
	ExpiredFiles int `json:"expiredFiles"`
}
