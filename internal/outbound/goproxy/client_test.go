package goproxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"gitlab.galvanize.com/cohorts/ent/sf/sdi-golang/sdi-go-2026-08-10/student-work/jiayang-wang/capstone/internal/campaign"
)

type cannedDoer struct {
	statusCode int
	body       string
	err        error
	request    *http.Request
}

func (d *cannedDoer) Do(req *http.Request) (*http.Response, error) {
	d.request = req
	if d.err != nil {
		return nil, d.err
	}
	return &http.Response{
		StatusCode: d.statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(d.body)),
		Header:     make(http.Header),
	}, nil
}

func TestClientLatest(t *testing.T) {
	fixture, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "goproxy_latest.json"))
	if err != nil {
		t.Fatalf("read provider fixture: %v", err)
	}
	doer := &cannedDoer{statusCode: http.StatusOK, body: string(fixture)}
	client := &Client{HTTP: doer, BaseURL: "https://proxy.example.test"}

	got, err := client.Latest(context.Background(), "github.com/spf13/cobra")
	if err != nil {
		t.Fatalf("Latest returned error: %v", err)
	}
	if got.LatestVersion != "v1.10.2" {
		t.Fatalf("LatestVersion = %q, want %q", got.LatestVersion, "v1.10.2")
	}
	if got.PublishedAt.IsZero() {
		t.Fatal("PublishedAt is zero")
	}
	if doer.request.Method != http.MethodGet {
		t.Fatalf("method = %q, want %q", doer.request.Method, http.MethodGet)
	}
	if doer.request.URL.String() != "https://proxy.example.test/github.com/spf13/cobra/@latest" {
		t.Fatalf("URL = %q", doer.request.URL.String())
	}
	if got := doer.request.Header.Get("Accept"); got != "application/json" {
		t.Fatalf("Accept = %q, want %q", got, "application/json")
	}
}

func TestClientLatestErrors(t *testing.T) {
	tests := []struct {
		name       string
		doer       *cannedDoer
		modulePath string
		wantErr    error
	}{
		{
			name:       "transport",
			doer:       &cannedDoer{err: errors.New("connection refused")},
			modulePath: "github.com/spf13/cobra",
			wantErr:    campaign.ErrProvider,
		},
		{
			name:       "non-200",
			doer:       &cannedDoer{statusCode: http.StatusNotFound},
			modulePath: "github.com/spf13/cobra",
			wantErr:    campaign.ErrProvider,
		},
		{
			name:       "malformed JSON",
			doer:       &cannedDoer{statusCode: http.StatusOK, body: "not-json"},
			modulePath: "github.com/spf13/cobra",
			wantErr:    campaign.ErrProvider,
		},
		{
			name:       "incomplete response",
			doer:       &cannedDoer{statusCode: http.StatusOK, body: `{"Version":"v1"}`},
			modulePath: "github.com/spf13/cobra",
			wantErr:    campaign.ErrProvider,
		},
		{
			name:       "invalid path segment",
			doer:       &cannedDoer{statusCode: http.StatusOK},
			modulePath: "github.com/../cobra",
			wantErr:    campaign.ErrInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{HTTP: tt.doer, BaseURL: "https://proxy.example.test"}
			_, err := client.Latest(context.Background(), tt.modulePath)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Latest error = %v, want errors.Is %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientLatestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	doer := &cannedDoer{err: fmt.Errorf("request context: %w", ctx.Err())}
	client := &Client{HTTP: doer, BaseURL: "https://proxy.example.test"}

	_, err := client.Latest(ctx, "github.com/spf13/cobra")
	if !errors.Is(err, campaign.ErrProvider) {
		t.Fatalf("Latest error = %v, want provider error", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Latest error = %v, want context cancellation", err)
	}
}
