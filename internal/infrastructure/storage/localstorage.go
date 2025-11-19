package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	UploadDir string
}

func NewLocalStorage(uploadDir string) Storage {
	// Đảm bảo thư mục tồn tại
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}
	return &LocalStorage{UploadDir: uploadDir}
}

func (s *LocalStorage) SaveFile(file *multipart.FileHeader, filename string) (string, error) {
	dst := filepath.Join(s.UploadDir, filename)

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return dst, nil
}

func (s *LocalStorage) DeleteFile(filename string) error {
	path := filepath.Join(s.UploadDir, filename) // Dòng này quan trọng
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		// Lỗi "The directory is not empty" xảy ra nếu path chỉ là "uploads"
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
