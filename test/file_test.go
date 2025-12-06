package test

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ==========================================
// HELPER FUNCTIONS
// ==========================================

// setupUserAndToken: Register -> Login -> Return (Token, Email)
func setupUserAndToken(t *testing.T) (string, string) {
	uniqueID := time.Now().UnixNano()
	username := fmt.Sprintf("user_%d", uniqueID)
	email := fmt.Sprintf("user_%d@example.com", uniqueID)
	password := "123456789"

	// 1. Register
	regBody := fmt.Sprintf(`{"username": "%s", "email": "%s", "password": "%s"}`, username, email, password)
	reqReg, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBufferString(regBody))
	reqReg.Header.Set("Content-Type", "application/json")

	recReg := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(recReg, reqReg)

	if recReg.Code != 200 && recReg.Code != 201 {
		t.Fatalf("Register failed: %v", recReg.Body.String())
	}

	// 2. Login
	loginBody := fmt.Sprintf(`{"email": "%s", "password": "%s"}`, email, password)
	reqLogin, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBufferString(loginBody))
	reqLogin.Header.Set("Content-Type", "application/json")

	recLogin := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(recLogin, reqLogin)
	assert.Equal(t, 200, recLogin.Code)

	// 3. Extract Token
	resp := ParseJSON(t, recLogin)
	token := ""

	if data, ok := resp["data"].(map[string]interface{}); ok {
		if t, ok := data["accessToken"].(string); ok {
			token = t
		}
	} else if t, ok := resp["accessToken"].(string); ok {
		token = t
	}

	if token == "" {
		t.Fatal("Cannot extract token from login response")
	}

	return token, email
}

// uploadFileForTest: Helper upload with full options
func uploadFileForTest(t *testing.T, token string, password string, availableFrom string, availableTo string, sharedWith string) (string, string) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormFile("file", "test_file.txt")
	io.WriteString(part, "Hello World Content")

	// Logic: Authenticated upload -> Private by default to test restrictions
	if token != "" {
		writer.WriteField("isPublic", "false")
	} else {
		writer.WriteField("isPublic", "true")
	}

	if password != "" {
		writer.WriteField("password", password)
		writer.WriteField("isPublic", "false") // Password implies private usually
	}
	if availableFrom != "" {
		writer.WriteField("availableFrom", availableFrom)
	}
	if availableTo != "" {
		writer.WriteField("availableTo", availableTo)
	}
	if sharedWith != "" {
		writer.WriteField("sharedWith", sharedWith)
	}

	writer.Close()

	req, _ := http.NewRequest("POST", "/api/files/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(rec, req)

	if rec.Code != 201 && rec.Code != 200 {
		t.Fatalf("Upload helper failed: %v", rec.Body.String())
	}

	resp := ParseJSON(t, rec)
	var fileData map[string]interface{}

	if data, ok := resp["data"].(map[string]interface{}); ok {
		fileData = data
	} else if val, ok := resp["file"].(map[string]interface{}); ok {
		fileData = val
	} else {
		fileData = resp
	}

	return fileData["id"].(string), fileData["shareToken"].(string)
}

// ==========================================
// TEST CASES
// ==========================================

func TestUpload_Scenarios(t *testing.T) {
	// Case: Anonymous Upload
	t.Run("Anonymous Upload", func(t *testing.T) {
		id, token := uploadFileForTest(t, "", "", "", "", "")
		assert.NotEmpty(t, id)
		assert.NotEmpty(t, token)
	})

	// Case: Auth Upload
	t.Run("Authenticated Upload", func(t *testing.T) {
		token, _ := setupUserAndToken(t)
		id, shareToken := uploadFileForTest(t, token, "", "", "", "")
		assert.NotEmpty(t, id)
		assert.NotEmpty(t, shareToken)
	})

	// Case: Bad Request (Missing File)
	t.Run("Missing File", func(t *testing.T) {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		writer.Close()
		req, _ := http.NewRequest("POST", "/api/files/upload", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)
		assert.Equal(t, 400, rec.Code)
	})
}

func TestDownload_PublicFile(t *testing.T) {
	token, _ := setupUserAndToken(t)
	_, shareToken := uploadFileForTest(t, "", "", "", "", "") // Anonymous file

	req, _ := http.NewRequest("GET", "/api/files/"+shareToken+"/download", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(rec, req)
	assert.Equal(t, 200, rec.Code)
}

func TestDownload_PasswordProtected(t *testing.T) {
	token, _ := setupUserAndToken(t)
	pass := "SecurePass123"
	_, shareToken := uploadFileForTest(t, "", pass, "", "", "")

	// Wrong password
	t.Run("Wrong Password", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/files/"+shareToken+"/download?password=wrong", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)
		assert.Equal(t, 403, rec.Code)
	})

	// Correct password
	t.Run("Correct Password", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/files/"+shareToken+"/download?password="+pass, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)
		assert.Equal(t, 200, rec.Code)
	})
}

func TestDownload_TimeRestricted(t *testing.T) {
	ownerToken, _ := setupUserAndToken(t)
	downloaderToken, downloaderEmail := setupUserAndToken(t)
	sharedJson := fmt.Sprintf(`["%s"]`, downloaderEmail)

	// Case 1: Locked File (Private + Shared + Future Start)
	t.Run("Locked File Shared User", func(t *testing.T) {
		future := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
		_, shareToken := uploadFileForTest(t, ownerToken, "", future, "", sharedJson)

		req, _ := http.NewRequest("GET", "/api/files/"+shareToken+"/download", nil)
		req.Header.Set("Authorization", "Bearer "+downloaderToken)

		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)
		assert.Equal(t, 423, rec.Code) // Expect Locked
	})

	// Case 2: Expired File (Private + Shared + Past End)
	t.Run("Expired File Shared User", func(t *testing.T) {
		pastFrom := time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339)
		pastTo := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
		_, shareToken := uploadFileForTest(t, ownerToken, "", pastFrom, pastTo, sharedJson)

		req, _ := http.NewRequest("GET", "/api/files/"+shareToken+"/download", nil)
		req.Header.Set("Authorization", "Bearer "+downloaderToken)

		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)
		// Check for error code (400 or 410 depending on implementation)
		assert.Contains(t, []int{400, 410, 422}, rec.Code)
	})

	// Case 3: Owner Bypass
	t.Run("Owner Bypass Time Check", func(t *testing.T) {
		future := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
		_, shareToken := uploadFileForTest(t, ownerToken, "", future, "", "")

		req, _ := http.NewRequest("GET", "/api/files/"+shareToken+"/download", nil)
		req.Header.Set("Authorization", "Bearer "+ownerToken) // Owner Token

		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)
		assert.Equal(t, 200, rec.Code)
	})
}

func TestDelete_Operations(t *testing.T) {
	// Anonymous
	t.Run("Anonymous Delete Fail", func(t *testing.T) {
		fileId, _ := uploadFileForTest(t, "", "", "", "", "")
		req, _ := http.NewRequest("DELETE", "/api/files/"+fileId, nil)
		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)
		assert.Contains(t, []int{401}, rec.Code)
	})

	// Owner
	t.Run("Owner Delete Success", func(t *testing.T) {
		token, _ := setupUserAndToken(t)
		fileId, _ := uploadFileForTest(t, token, "", "", "", "")

		req, _ := http.NewRequest("DELETE", "/api/files/"+fileId, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)
		assert.Equal(t, 200, rec.Code)
	})

	// Other User
	t.Run("Attacker Delete Fail", func(t *testing.T) {
		ownerToken, _ := setupUserAndToken(t)
		attackerToken, _ := setupUserAndToken(t)
		fileId, _ := uploadFileForTest(t, ownerToken, "", "", "", "")

		req, _ := http.NewRequest("DELETE", "/api/files/"+fileId, nil)
		req.Header.Set("Authorization", "Bearer "+attackerToken)
		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)
		assert.Equal(t, 403, rec.Code)
	})
}

func TestMyFiles_List(t *testing.T) {
	token, _ := setupUserAndToken(t)
	uploadFileForTest(t, token, "", "", "", "")
	uploadFileForTest(t, token, "", "", "", "")

	req, _ := http.NewRequest("GET", "/api/files/my?page=1&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)
	resp := ParseJSON(t, rec)
	assert.NotNil(t, resp["files"])
	files := resp["files"].([]interface{})
	assert.GreaterOrEqual(t, len(files), 2)
}
