package storage

import (
	// Import errors
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	UploadDir string
}

func NewLocalStorage(uploadDir string) Storage {
	// LƯU Ý QUAN TRỌNG:
	// Nếu bạn biết thư mục 'uploads' luôn nằm trong 'cmd/server',
	// và bạn chạy chương trình từ thư mục gốc dự án, bạn cần đảm bảo
	// uploadDir được thiết lập thành "cmd/server/uploads" trong cấu hình.
	// Nếu uploadDir chỉ là "uploads", hãy cố gắng biến nó thành đường dẫn tuyệt đối.

	absPath, err := filepath.Abs(uploadDir)
	if err != nil {
		// Log hoặc panic nếu không thể lấy đường dẫn tuyệt đối
		fmt.Printf("Warning: Failed to get absolute path for %s. Using relative path.\n", uploadDir)
		absPath = uploadDir
	}

	// Đảm bảo thư mục tồn tại
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		os.MkdirAll(absPath, 0755)
	}
	// Sử dụng đường dẫn tuyệt đối đã được tính toán
	return &LocalStorage{UploadDir: absPath}
}

func (s *LocalStorage) SaveFile(file *multipart.FileHeader, filename string) (string, error) {
	// Dòng này sử dụng đường dẫn tuyệt đối đã được lưu trong s.UploadDir
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
	path := filepath.Clean(filepath.Join(s.UploadDir, filename))
	if filename == "" {
		log.Printf("[DEBUG DELETE] Attempting to delete file:\n- Filename (DB): %s\n- UploadDir: %s\n- Full Path: %s", filename, s.UploadDir, path)
		return nil
	}

	// >>>>>> DÒNG DEBUG QUAN TRỌNG <<<<<<
	// Dùng log package nếu bạn có, hoặc tạm dùng fmt.Println
	log.Printf("[DEBUG DELETE] Attempting to delete file:\n- Filename (DB): %s\n- UploadDir: %s\n- Full Path: %s", filename, s.UploadDir, path)
	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

	err := os.Remove(path)

	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		// Lỗi xóa file thông thường
		return fmt.Errorf("failed to delete file %s at %s: %w", filename, path, err)
	}
	return nil
}
