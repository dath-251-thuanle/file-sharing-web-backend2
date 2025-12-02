package storage

import (
	"io"
	"mime/multipart"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
)

type Storage interface {
	SaveFile(file *multipart.FileHeader, filename string) (string, *utils.ReturnStatus)
	DeleteFile(filename string) *utils.ReturnStatus
	GetFile(filename string) (io.Reader, *utils.ReturnStatus) // Cần cho Download
}
