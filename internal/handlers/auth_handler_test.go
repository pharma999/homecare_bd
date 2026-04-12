package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"home_care_backend/internal/config"
	"home_care_backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
	// Minimal config so JWT utils work without a real .env
	config.AppConfig = &config.Config{
		JWTSecret:        "test-secret",
		JWTExpiryHours:   24,
		OTPExpiryMinutes: 5,
		AppName:          "HomeCareTest",
		DevOTPBypass:     true,
	}
}

func newRouter() *gin.Engine {
	r := gin.New()
	return r
}

// ── /auth/send-otp ───────────────────────────────────────────────────────────

func TestLogin_MissingPhone(t *testing.T) {
	r := newRouter()
	r.POST("/auth/send-otp", handlers.Login)

	body, _ := json.Marshal(map[string]string{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/send-otp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLogin_ShortPhone(t *testing.T) {
	r := newRouter()
	r.POST("/auth/send-otp", handlers.Login)

	body, _ := json.Marshal(map[string]string{"phone_number": "123"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/send-otp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ── /auth/verify-otp ────────────────────────────────────────────────────────

func TestVerifyOTP_MissingFields(t *testing.T) {
	r := newRouter()
	r.POST("/auth/verify-otp", handlers.VerifyOTP)

	body, _ := json.Marshal(map[string]string{"phone_number": "9876543210"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/verify-otp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
