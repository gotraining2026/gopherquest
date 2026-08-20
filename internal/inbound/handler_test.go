package inbound

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gitlab.galvanize.com/cohorts/ent/sf/sdi-golang/sdi-go-2026-08-10/student-work/jiayang-wang/capstone/internal/campaign"
)

type fakeStore struct {
	items  map[int64]campaign.Campaign
	nextID int64
}

func newFakeStore() *fakeStore {
	return &fakeStore{items: make(map[int64]campaign.Campaign), nextID: 1}
}

func (f *fakeStore) Create(_ context.Context, input campaign.NewCampaign) (campaign.Campaign, error) {
	now := time.Date(2026, time.August, 20, 12, 0, 0, 0, time.UTC)
	item := campaign.Campaign{
		ID:         f.nextID,
		PlayerName: input.PlayerName,
		Title:      input.Title,
		ModulePath: input.ModulePath,
		Difficulty: input.Difficulty,
		Stage:      campaign.StageModule,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	f.items[item.ID] = item
	f.nextID++
	return item, nil
}

func (f *fakeStore) List(context.Context) ([]campaign.Campaign, error) {
	items := make([]campaign.Campaign, 0, len(f.items))
	for _, item := range f.items {
		items = append(items, item)
	}
	return items, nil
}

func (f *fakeStore) Get(_ context.Context, id int64) (campaign.Campaign, error) {
	item, ok := f.items[id]
	if !ok {
		return campaign.Campaign{}, campaign.ErrNotFound
	}
	return item, nil
}

func (f *fakeStore) Update(_ context.Context, id int64, input campaign.UpdateCampaign) (campaign.Campaign, error) {
	item, err := f.Get(context.Background(), id)
	if err != nil {
		return campaign.Campaign{}, err
	}
	if item.Stage != campaign.StageModule {
		return campaign.Campaign{}, campaign.ErrWrongStage
	}
	item.PlayerName = input.PlayerName
	item.Title = input.Title
	item.ModulePath = input.ModulePath
	item.Difficulty = input.Difficulty
	f.items[id] = item
	return item, nil
}

func (f *fakeStore) Delete(_ context.Context, id int64) error {
	if _, ok := f.items[id]; !ok {
		return campaign.ErrNotFound
	}
	delete(f.items, id)
	return nil
}

func (f *fakeStore) Advance(_ context.Context, id int64, expected, next campaign.Stage, points int) (campaign.Campaign, error) {
	item, err := f.Get(context.Background(), id)
	if err != nil {
		return campaign.Campaign{}, err
	}
	if item.Stage != expected {
		return campaign.Campaign{}, campaign.ErrWrongStage
	}
	item.Stage = next
	item.Score += points
	f.items[id] = item
	return item, nil
}

type fakeModuleLookup struct {
	info campaign.ModuleInfo
}

func (f fakeModuleLookup) Latest(context.Context, string) (campaign.ModuleInfo, error) {
	return f.info, nil
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	service := campaign.NewService(
		newFakeStore(),
		fakeModuleLookup{info: campaign.ModuleInfo{
			LatestVersion: "v1.10.2",
			PublishedAt:   time.Date(2025, time.December, 3, 23, 51, 15, 0, time.UTC),
		}},
	)
	handler := NewHandler(service)
	mux, err := Mux(handler)
	if err != nil {
		t.Fatalf("build mux: %v", err)
	}
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return ts
}

func TestHealthz(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read health response: %v", err)
	}
	if resp.StatusCode != http.StatusOK || string(body) != "ok" {
		t.Fatalf("health response = %d %q", resp.StatusCode, body)
	}
}

func TestCRUDRoutes(t *testing.T) {
	ts := newTestServer(t)
	client := ts.Client()

	createBody := `{"playerName":"Jiayang","title":"Protocol Run","modulePath":"github.com/spf13/cobra","difficulty":"beginner"}`
	resp := doRequest(t, client, http.MethodPost, ts.URL+"/campaigns", createBody)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	if created["id"] != float64(1) || created["stage"] != string(campaign.StageModule) {
		t.Fatalf("created response = %#v", created)
	}

	resp = doRequest(t, client, http.MethodGet, ts.URL+"/campaigns", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var listed []map[string]any
	decodeJSON(t, resp, &listed)
	if len(listed) != 1 {
		t.Fatalf("listed campaigns = %#v", listed)
	}

	resp = doRequest(t, client, http.MethodGet, ts.URL+"/campaigns/1", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var got map[string]any
	decodeJSON(t, resp, &got)
	moduleInfo, ok := got["moduleInfo"].(map[string]any)
	if !ok || moduleInfo["latestVersion"] != "v1.10.2" {
		t.Fatalf("enriched response = %#v", got)
	}

	updateBody := `{"playerName":"Updated","title":"Updated Run","modulePath":"github.com/spf13/cobra","difficulty":"advanced"}`
	resp = doRequest(t, client, http.MethodPut, ts.URL+"/campaigns/1", updateBody)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()

	resp = doRequest(t, client, http.MethodDelete, ts.URL+"/campaigns/1", "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	resp.Body.Close()
}

func TestCRUDErrors(t *testing.T) {
	ts := newTestServer(t)
	client := ts.Client()

	resp := doRequest(t, client, http.MethodGet, ts.URL+"/campaigns/999", "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("missing campaign status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
	resp.Body.Close()

	invalidBody := `{"playerName":"","title":"","modulePath":"invalid","difficulty":"beginner"}`
	resp = doRequest(t, client, http.MethodPost, ts.URL+"/campaigns", invalidBody)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid campaign status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	resp.Body.Close()
}

func TestStatusForError(t *testing.T) {
	tests := []struct {
		err    error
		status int
	}{
		{err: campaign.ErrInvalid, status: http.StatusBadRequest},
		{err: campaign.ErrNotFound, status: http.StatusNotFound},
		{err: campaign.ErrWrongStage, status: http.StatusConflict},
		{err: campaign.ErrProvider, status: http.StatusBadGateway},
		{err: errors.New("unexpected"), status: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		if got := statusForError(tt.err); got != tt.status {
			t.Errorf("statusForError(%v) = %d, want %d", tt.err, got, tt.status)
		}
	}
}

func doRequest(t *testing.T, client *http.Client, method, url, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response, target any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
