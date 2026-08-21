package inbound

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type requestIDKey struct{}

// RequestLogger records one structured completion event for each request.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = newRequestID()
		}

		ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
		w.Header().Set("X-Request-ID", requestID)
		recorder := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r.WithContext(ctx))

		status := recorder.statusCode()
		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"operation", operationFor(r),
			"status", status,
			"duration", time.Since(start),
			"request_id", requestID,
			"error_class", errorClass(status),
		}
		if id, ok := campaignIDFor(r); ok {
			attrs = append(attrs, "campaign_id", id)
		}

		level := slog.LevelInfo
		switch {
		case status >= http.StatusInternalServerError:
			level = slog.LevelError
		case status >= http.StatusBadRequest:
			level = slog.LevelWarn
		}
		slog.Log(ctx, level, "http request complete", attrs...)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(body)
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func (r *statusRecorder) statusCode() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}

func newRequestID() string {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err == nil {
		return "req-" + hex.EncodeToString(raw[:])
	}
	return "req-" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

func operationFor(r *http.Request) string {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 1 && parts[0] == "campaigns" {
		switch r.Method {
		case http.MethodGet:
			return "listCampaigns"
		case http.MethodPost:
			return "createCampaign"
		}
	}
	if len(parts) == 2 && parts[0] == "campaigns" {
		switch r.Method {
		case http.MethodGet:
			return "getCampaign"
		case http.MethodPut:
			return "updateCampaign"
		case http.MethodDelete:
			return "deleteCampaign"
		}
	}
	if len(parts) == 3 && parts[0] == "campaigns" && parts[2] == "attempts" && r.Method == http.MethodPost {
		return "submitAttempt"
	}
	if r.Method == http.MethodGet && r.URL.Path == "/healthz" {
		return "healthz"
	}
	return ""
}

func campaignIDFor(r *http.Request) (int64, bool) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != "campaigns" {
		return 0, false
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	return id, err == nil
}

func errorClass(status int) string {
	switch {
	case status == http.StatusBadGateway:
		return "provider"
	case status >= http.StatusInternalServerError:
		return "internal"
	case status == http.StatusConflict:
		return "conflict"
	case status == http.StatusNotFound:
		return "not_found"
	case status >= http.StatusBadRequest:
		return "validation"
	default:
		return ""
	}
}
