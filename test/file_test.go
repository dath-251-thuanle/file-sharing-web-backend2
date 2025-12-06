package test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileUpload(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	tmp, _ := os.CreateTemp("", "upload*.txt")
	tmp.WriteString("hello world")
	tmp.Seek(0, 0)

	part, _ := writer.CreateFormFile("file", "hello.txt")
	io.Copy(part, tmp)
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/files/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(rec, req)

	assert.Contains(t, []int{201, 200}, rec.Code)
}