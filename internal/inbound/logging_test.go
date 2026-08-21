package inbound

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"
)

func TestRequestLogger(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	req := httptest.NewRequest(http.MethodGet, "/campaigns/42", nil)
	req.Header.Set("X-Request-ID", "demo-request")
	rec := httptest.NewRecorder()

	RequestLogger(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if got := rec.Header().Get("X-Request-ID"); got != "demo-request" {
		t.Fatalf("response request ID = %q", got)
	}

	var event map[string]any
	if err := json.Unmarshal(logs.Bytes(), &event); err != nil {
		t.Fatalf("decode log event: %v", err)
	}
	checks := map[string]any{
		"msg":         "http request complete",
		"method":      "GET",
		"path":        "/campaigns/42",
		"operation":   "getCampaign",
		"status":      float64(http.StatusNotFound),
		"request_id":  "demo-request",
		"campaign_id": float64(42),
		"error_class": "not_found",
		"level":       "WARN",
	}
	for key, want := range checks {
		if got := event[key]; got != want {
			t.Errorf("log[%q] = %#v, want %#v", key, got, want)
		}
	}
	if _, ok := event["duration"]; !ok {
		t.Error("log has no duration")
	}
}

func TestRequestLoggerGeneratesRequestID(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodDelete, "/campaigns/7", nil)
	rec := httptest.NewRecorder()

	RequestLogger(next).ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); len(got) < len("req-")+1 || got[:len("req-")] != "req-" {
		t.Fatalf("generated request ID = %q", got)
	}
}

func TestErrorClass(t *testing.T) {
	tests := []struct {
		status int
		want   string
	}{
		{status: http.StatusOK, want: ""},
		{status: http.StatusBadRequest, want: "validation"},
		{status: http.StatusNotFound, want: "not_found"},
		{status: http.StatusConflict, want: "conflict"},
		{status: http.StatusBadGateway, want: "provider"},
		{status: http.StatusInternalServerError, want: "internal"},
	}
	for _, tt := range tests {
		if got := errorClass(tt.status); got != tt.want {
			t.Errorf("errorClass(%d) = %q, want %q", tt.status, got, tt.want)
		}
	}
}
