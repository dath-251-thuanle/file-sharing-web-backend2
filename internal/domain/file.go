package domain

import "time"

type File struct {
	Id            string     `json:"id" db:"id"`
	OwnerId       *string    `json:"ownerId" db:"user_id"`          // Tên cột là user_id
	FileName      string     `json:"fileName" db:"name"`            // Tên cột là name
	StorageName   string     `json:"storageName" db:"storage_name"` // (Giả định cột này tồn tại)
	FileSize      int64      `json:"fileSize" db:"size"`
	MimeType      string     `json:"mimeType" db:"type"` // Tên cột là type
	ShareToken    string     `json:"shareToken" db:"share_token"`
	IsPublic      bool       `json:"isPublic" db:"is_public"`       // (Giả định cột này tồn tại)
	HasPassword   bool       `json:"hasPassword" db:"has_password"` // (Giả định cột này tồn tại)
	PasswordHash  *string    `json:"-" db:"password"`               // Tên cột là password
	EnableTOTP    bool       `json:"enableTOTP" db:"enable_totp"`
	AvailableFrom time.Time  `json:"availableFrom" db:"available_from"`
	AvailableTo   time.Time  `json:"availableTo" db:"available_to"`
	ValidityDays  int        `json:"validityDays" db:"validity_days"` // (Giả định cột này tồn tại)
	Status        string     `json:"status"`                          // pending, active, expired (Không lưu DB)
	CreatedAt     time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt     *time.Time `json:"updatedAt" db:"updated_at"` // (Giả định cột này tồn tại)
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
