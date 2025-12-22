package test

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuth_TOTP_Flow(t *testing.T) {
	ResetDB(t)
	t.Cleanup(func() { ResetDB(t) })

	var (
		testEmail    = fmt.Sprintf("testuser_%d@example.com", time.Now().UnixNano())
		testPassword = "Password123"
		testUsername = "testuser"

		authToken  string
		totpSecret string
	)

	// ---------------------------
	// 1. REGISTER
	// ---------------------------
	t.Run("REGISTER", func (t *testing.T) {
		body := fmt.Sprintf(`{
			"username": "%s",
			"email": "%s",
			"password": "%s"
		}`, testUsername, testEmail, testPassword)

		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)

		json := ParseJSON(t, rec)
		assert.Equal(t, "User registered successfully", json["message"])
	})

	// ---------------------------
	// 2. LOGIN (no TOTP)
	// ---------------------------
	t.Run("LOGIN", func (t *testing.T) {

		body := fmt.Sprintf(`{
			"email": "%s",
			"password": "%s"
		}`, testEmail, testPassword)

		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)

		json := ParseJSON(t, rec)
		assert.NotEmpty(t, json["accessToken"])

		authToken = json["accessToken"].(string)
	})

	// ---------------------------
	// 3. SETUP TOTP
	// ---------------------------
	t.Run("SETUP TOTP", func (t *testing.T) {


		req := httptest.NewRequest("POST", "/auth/totp/setup", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)

		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)

		json := ParseJSON(t, rec)
		setup := json["totpSetup"].(map[string]interface{})

		totpSecret = setup["secret"].(string)

		assert.NotEmpty(t, totpSecret)
		assert.NotEmpty(t, setup["qrCode"])
	})

	// ---------------------------
	// 4. MANUAL ENABLE TOTP IN DB
	// ---------------------------
	t.Run("MANUAL ENABLE TOTP IN DB", func (t *testing.T) {

		db := TestApp.DB()
		_, err := db.Exec(`UPDATE users SET enabletotp = true WHERE email=$1`, testEmail)

		assert.NoError(t, err)
	})

	// ---------------------------
	// 5. LOGIN AGAIN -> requireTOTP
	// ---------------------------
	t.Run("MANUAL ENABLE TOTP IN DB", func (t *testing.T) {

		body := fmt.Sprintf(`{
			"email": "%s",
			"password": "%s"
		}`, testEmail, testPassword)

		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)

		json := ParseJSON(t, rec)

		assert.Equal(t, true, json["requireTOTP"])
		assert.NotEmpty(t, json["cid"])

		serverSecret := os.Getenv("JWT_SECRET_KEY")
		assert.NotEmpty(t, serverSecret)
	})

	// ---------------------------
	// 6. GET PROFILE (token cũ vẫn valid)
	// ---------------------------
	t.Run("GET PROFILE", func (t *testing.T) {

		req := httptest.NewRequest("GET", "/user", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)

		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)
	})

	// ---------------------------
	// 7. LOGOUT
	// ---------------------------
	t.Run("LOGOUT", func (t *testing.T) {

		req := httptest.NewRequest("POST", "/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)

		rec := httptest.NewRecorder()
		TestApp.Router().ServeHTTP(rec, req)

		assert.Equal(t, 200, rec.Code)
	})
}