package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==========================================
// ADMIN HELPER
// ==========================================

// setupAdminToken: Tạo user -> Update DB -> Login LẠI -> Trả về Token mới
func setupAdminToken(t *testing.T) string {
	// 1. Tạo user và lấy email (Token cũ vứt đi vì nó chỉ có quyền user)
	_, email := setupUserAndToken(t)

	// 2. Thăng cấp lên Admin trong DB
	db := TestApp.DB()
	if db == nil {
		t.Fatal("Database connection is nil")
	}

	// [POSTGRES] Dùng $1
	_, err := db.Exec("UPDATE users SET role = 'admin' WHERE email = $1", email)
	if err != nil {
		t.Fatalf("Failed to promote user to admin: %v", err)
	}

	// 3. [QUAN TRỌNG] Login LẠI để lấy Token MỚI chứa quyền Admin
	// Password mặc định trong setupUserAndToken là "123456789"
	loginBody := fmt.Sprintf(`{"email": "%s", "password": "%s"}`, email, "123456789")
	reqLogin, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBufferString(loginBody))
	reqLogin.Header.Set("Content-Type", "application/json")

	recLogin := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(recLogin, reqLogin)

	assert.Equal(t, 200, recLogin.Code, "Admin re-login failed")

	// 4. Trả về Token MỚI
	resp := ParseJSON(t, recLogin)

	// Logic lấy token (copy từ setupUserAndToken)
	if data, ok := resp["data"].(map[string]interface{}); ok {
		if token, ok := data["accessToken"].(string); ok {
			return token
		}
	}
	if token, ok := resp["accessToken"].(string); ok {
		return token
	}

	t.Fatal("Cannot extract new admin token")
	return ""
}

// ==========================================
// TEST CASES: SYSTEM POLICY
// ==========================================

func TestAdmin_GetPolicy(t *testing.T) {
	t.Run("Admin Get Policy Success", func(t *testing.T) {
		adminToken := setupAdminToken(t)

		req, _ := http.NewRequest("GET", "/api/admin/policy", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)

		resp := ParseJSON(t, rec)

		var data map[string]interface{}
		if d, ok := resp["data"].(map[string]interface{}); ok {
			data = d
		} else {
			data = resp
		}

		// [FIX] Sửa key thành PascalCase (Viết hoa chữ đầu) theo đúng log server trả về
		assert.NotNil(t, data["MaxFileSizeMB"])
		assert.NotNil(t, data["DefaultValidityDays"])
	})

	t.Run("User Get Policy Forbidden", func(t *testing.T) {
		userToken, _ := setupUserAndToken(t)

		req, _ := http.NewRequest("GET", "/api/admin/policy", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)

		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)

		assert.Equal(t, 403, rec.Code)
	})
}

func TestAdmin_UpdatePolicy(t *testing.T) {
	adminToken := setupAdminToken(t)

	t.Run("Update Policy Success", func(t *testing.T) {
		// [FIX] Key gửi lên cũng nên để PascalCase nếu server dùng struct binding mặc định
		body := map[string]interface{}{
			"MaxFileSizeMB":       100,
			"DefaultValidityDays": 7,
			"MaxValidityDays":     30,
			"AllowedMimeTypes":    []string{"image/png", "image/jpeg"},
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("PATCH", "/api/admin/policy", bytes.NewBuffer(jsonBody))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)

		// Verify update
		reqGet, _ := http.NewRequest("GET", "/api/admin/policy", nil)
		reqGet.Header.Set("Authorization", "Bearer "+adminToken)
		recGet := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(recGet, reqGet)

		resp := ParseJSON(t, recGet)
		var data map[string]interface{}
		if d, ok := resp["data"].(map[string]interface{}); ok {
			data = d
		} else {
			data = resp
		}

		// [FIX] Sửa key thành PascalCase
		assert.Equal(t, 100.0, data["MaxFileSizeMB"])
	})

	t.Run("Update Policy Invalid Data", func(t *testing.T) {
		body := map[string]interface{}{
			"MaxFileSizeMB": -10,
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("PATCH", "/api/admin/policy", bytes.NewBuffer(jsonBody))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)

		assert.Equal(t, 400, rec.Code)
	})
}

// ==========================================
// TEST CASES: CLEANUP
// ==========================================

func TestAdmin_Cleanup(t *testing.T) {
	adminToken := setupAdminToken(t)

	t.Run("Trigger Cleanup Success", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/admin/cleanup", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)
	})

	t.Run("User Cannot Cleanup", func(t *testing.T) {
		userToken, _ := setupUserAndToken(t)
		req, _ := http.NewRequest("POST", "/api/admin/cleanup", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)

		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)

		assert.Equal(t, 403, rec.Code)
	})
}
