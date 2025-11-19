package storage

import (
	"mime/multipart"
)

type Storage interface {
	SaveFile(file *multipart.FileHeader, filename string) (string, error)
	DeleteFile(filename string) error
	// GetFile(filename string) (io.Reader, error) // Cần cho Download
}
