package storage

import (
	// Import errors
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
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

func (s *LocalStorage) SaveFile(file *multipart.FileHeader, filename string) (string, *utils.ReturnStatus) {
	// Dòng này sử dụng đường dẫn tuyệt đối đã được lưu trong s.UploadDir
	dst := filepath.Join(s.UploadDir, filename)

	src, err := file.Open()
	if err != nil {
		return "", utils.ResponseMsg(utils.ErrCodeInternal, fmt.Sprintf("failed to open file: %s", err))
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return "", utils.ResponseMsg(utils.ErrCodeInternal, fmt.Sprintf("failed to create destination file: %s", err))
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	if err != nil {
		return "", utils.ResponseMsg(utils.ErrCodeInternal, fmt.Sprintf("failed to save file: %s", err))
	}

	return dst, nil
}

func (s *LocalStorage) GetFile(filename string) (io.Reader, *utils.ReturnStatus) {
	dst := filepath.Join(s.UploadDir, filename)

	file, err := os.Open(dst)
	if err != nil {
		return nil, utils.ResponseMsg(utils.ErrCodeInternal, fmt.Sprintf("failed to open file: %s", err))
	}

	var reader io.Reader = file

	return reader, nil
}

func (s *LocalStorage) DeleteFile(filename string) *utils.ReturnStatus {
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
			return utils.Response(utils.ErrCodeFileNotFound)
		}

		// Lỗi xóa file thông thường
		return utils.ResponseMsg(utils.ErrCodeInternal, fmt.Sprintf("failed to delete file %s at %s: %s", filename, path, err))
	}
	return nil
}
