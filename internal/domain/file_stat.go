package domain

import "time"

type FileStat struct {
	Id                 string
	FileId             string
	FileName           string
	UserDownloadCount  int
	TotalDownloadCount int
	LastDownloadedAt   time.Time
	CreatedAt          time.Time
}
