package domain

import "time"

type SharedWith struct {
	Id         string    `json:"id" db:"id"`
	FileId     string    `json:"fileId" db:"file_id"`
	UserId     string    `json:"userId" db:"user_id"`
	SharedAt   time.Time `json:"sharedAt" db:"shared_at"`
	Permission string    `json:"permission" db:"permission"` // Ví dụ: read, write
}
