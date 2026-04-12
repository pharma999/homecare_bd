package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"home_care_backend/internal/handlers"
)

// ── UpdateUserProfile ─────────────────────────────────────────────────────────

func TestUpdateUserProfile_WrongUser(t *testing.T) {
	r := newRouter()
	// Authenticated as user-2 but trying to update user-1's profile
	r.POST("/user/:userId/update", injectUser("user-2", "PATIENT"), handlers.UpdateUserProfile)

	body, _ := json.Marshal(map[string]string{"name": "Alice"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/user/user-1/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestUpdateUserProfile_InvalidBody(t *testing.T) {
	r := newRouter()
	r.POST("/user/:userId/update", injectUser("user-1", "PATIENT"), handlers.UpdateUserProfile)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/user/user-1/update", bytes.NewReader([]byte(`{bad json`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ── UpdateUserAddress ─────────────────────────────────────────────────────────

func TestUpdateUserAddress_WrongUser(t *testing.T) {
	r := newRouter()
	r.POST("/user/:userId/address/update", injectUser("user-2", "PATIENT"), handlers.UpdateUserAddress)

	body, _ := json.Marshal(map[string]string{
		"addressType": "address1",
		"street":      "MG Road",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/user/user-1/address/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestUpdateUserAddress_MissingAddressType(t *testing.T) {
	r := newRouter()
	r.POST("/user/:userId/address/update", injectUser("user-1", "PATIENT"), handlers.UpdateUserAddress)

	body, _ := json.Marshal(map[string]string{
		"street": "MG Road",
		// addressType is required
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/user/user-1/address/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Missing addressType → 400 or 404 (user not found in DB)
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("expected 400 or 404, got %d — body: %s", w.Code, w.Body.String())
	}
}

// ── DeleteUserAccount ─────────────────────────────────────────────────────────

func TestDeleteUserAccount_WrongUser(t *testing.T) {
	r := newRouter()
	r.POST("/user/:userId/delete", injectUser("user-2", "PATIENT"), handlers.DeleteUserAccount)

	body, _ := json.Marshal(map[string]string{"userId": "user-1"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/user/user-1/delete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

// ── AddFamilyMember ───────────────────────────────────────────────────────────

func TestAddFamilyMember_MissingFields(t *testing.T) {
	r := newRouter()
	r.POST("/family/members", injectUser("user-1", "PATIENT"), handlers.AddFamilyMember)

	body, _ := json.Marshal(map[string]string{
		"family_user_id": "user-2",
		// relation is required
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/family/members", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d — body: %s", w.Code, w.Body.String())
	}
}

// ── Notifications ─────────────────────────────────────────────────────────────

func TestMarkNotificationRead_NoDBConnection(t *testing.T) {
	r := newRouter()
	r.PATCH("/notifications/:notifId/read", injectUser("user-1", "PATIENT"), handlers.MarkNotificationRead)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/notifications/notif-999/read", nil)
	r.ServeHTTP(w, req)

	// Without DB: expect 404 (MatchedCount == 0)
	if w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("expected 404 or 500, got %d — body: %s", w.Code, w.Body.String())
	}
}
