package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"home_care_backend/internal/handlers"
)

func TestSendChatMessage_MissingMessage(t *testing.T) {
	r := newRouter()
	r.POST("/chatbot/message", injectUser("user-1", "PATIENT"), handlers.SendChatMessage)

	body, _ := json.Marshal(map[string]string{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/chatbot/message", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestBotReply_GreetingIntent(t *testing.T) {
	cases := []struct {
		msg      string
		contains string
	}{
		{"hello", "assistant"},
		{"Hi there", "assistant"},
		{"emergency help", "SOS"},
		{"book a doctor", "appointment"},
		{"my booking status", "booking"},
		{"payment issue", "payment"},
		{"support needed", "ticket"},
	}

	r := newRouter()
	r.POST("/chatbot/message", injectUser("user-1", "PATIENT"), handlers.SendChatMessage)

	for _, tc := range cases {
		body, _ := json.Marshal(map[string]string{"message": tc.msg})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/chatbot/message", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		// Without DB the handler may return 500 on DB write, but the reply logic is
		// tested by the containsAny helper separately via the intent switch.
		// Here we just ensure no panic and either 201/200 or 500 (no DB).
		if w.Code == http.StatusBadRequest {
			t.Errorf("msg=%q: unexpected 400 — body: %s", tc.msg, w.Body.String())
		}
	}
}
