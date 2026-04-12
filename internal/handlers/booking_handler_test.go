package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"home_care_backend/internal/handlers"
	"home_care_backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// injectUser adds a fake authenticated user into the Gin context.
func injectUser(userID, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

// ── Create Booking ───────────────────────────────────────────────────────────

func TestCreateBooking_MissingServiceID(t *testing.T) {
	r := newRouter()
	r.POST("/bookings", injectUser("user-1", "PATIENT"), handlers.CreateBooking)

	body, _ := json.Marshal(map[string]string{
		"booking_type":   "QUICK",
		"patient_address": "123 Main St",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// service_id is required — expects 400 or 404 (DB call for non-existent service)
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("expected 400 or 404, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestCreateBooking_MissingPatientAddress(t *testing.T) {
	r := newRouter()
	r.POST("/bookings", injectUser("user-1", "PATIENT"), handlers.CreateBooking)

	body, _ := json.Marshal(map[string]string{
		"service_id":   "svc-123",
		"booking_type": "QUICK",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ── Cart ─────────────────────────────────────────────────────────────────────

func TestAddToCart_MissingFields(t *testing.T) {
	r := newRouter()
	r.POST("/cart/add", injectUser("user-1", "PATIENT"), handlers.AddToCart)

	body, _ := json.Marshal(map[string]string{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cart/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateCartQuantity_InvalidBody(t *testing.T) {
	r := newRouter()
	r.POST("/cart/update-quantity", injectUser("user-1", "PATIENT"), handlers.UpdateCartQuantity)

	body := bytes.NewReader([]byte(`not-json`))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cart/update-quantity", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ── Payment ───────────────────────────────────────────────────────────────────

func TestInitiatePayment_MissingAmount(t *testing.T) {
	r := newRouter()
	r.POST("/payments/initiate", injectUser("user-1", "PATIENT"), handlers.InitiatePayment)

	body, _ := json.Marshal(map[string]string{
		"payment_method": "UPI",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/payments/initiate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ── Middleware helper used across tests ───────────────────────────────────────

// Ensure middleware package is used (avoids unused import error in test files
// that import only for side effects).
var _ = middleware.AuthRequired
