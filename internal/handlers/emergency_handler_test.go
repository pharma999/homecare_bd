package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"home_care_backend/internal/handlers"
)

// ── TriggerSOS ───────────────────────────────────────────────────────────────

func TestTriggerSOS_MissingLocation(t *testing.T) {
	r := newRouter()
	r.POST("/emergency/sos", injectUser("user-1", "PATIENT"), handlers.TriggerSOS)

	// patient_latitude and patient_longitude are required
	body, _ := json.Marshal(map[string]string{
		"symptom_description": "chest pain",
		"patient_address":     "123 Main St",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/emergency/sos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Missing required fields → 400
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestTriggerSOS_ValidPayload_NoDBConnection(t *testing.T) {
	r := newRouter()
	r.POST("/emergency/sos", injectUser("user-1", "PATIENT"), handlers.TriggerSOS)

	body, _ := json.Marshal(map[string]string{
		"patient_latitude":    "26.8467",
		"patient_longitude":   "80.9462",
		"patient_address":     "123 Main St, Lucknow",
		"symptom_description": "severe chest pain",
		"emergency_type":      "CARDIAC",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/emergency/sos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Without a DB connection this panics or returns 500/201.
	// We just verify the handler parses the request without crashing on bad JSON.
	if w.Code == http.StatusBadRequest {
		t.Errorf("unexpected 400 — body: %s", w.Body.String())
	}
}

// ── Medical Records ───────────────────────────────────────────────────────────

func TestUploadMedicalRecord_MissingFields(t *testing.T) {
	r := newRouter()
	r.POST("/medical-records", injectUser("user-1", "PATIENT"), handlers.UploadMedicalRecord)

	body, _ := json.Marshal(map[string]string{
		"title": "Blood Test",
		// record_type and record_date are required
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/medical-records", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestUploadMedicalRecord_InvalidDate(t *testing.T) {
	r := newRouter()
	r.POST("/medical-records", injectUser("user-1", "PATIENT"), handlers.UploadMedicalRecord)

	body, _ := json.Marshal(map[string]string{
		"record_type": "LAB_REPORT",
		"title":       "CBC",
		"record_date": "not-a-date",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/medical-records", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid date, got %d — body: %s", w.Code, w.Body.String())
	}
}
