package campaign

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeStore struct {
	created NewCampaign
	item    Campaign
	items   []Campaign
	err     error
}

func (f *fakeStore) Create(_ context.Context, input NewCampaign) (Campaign, error) {
	f.created = input
	if f.err != nil {
		return Campaign{}, f.err
	}
	return f.item, nil
}

func (f *fakeStore) List(context.Context) ([]Campaign, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func (f *fakeStore) Get(context.Context, int64) (Campaign, error) {
	if f.err != nil {
		return Campaign{}, f.err
	}
	return f.item, nil
}

func (f *fakeStore) Update(context.Context, int64, UpdateCampaign) (Campaign, error) {
	if f.err != nil {
		return Campaign{}, f.err
	}
	return f.item, nil
}

func (f *fakeStore) Delete(context.Context, int64) error {
	return f.err
}

func (f *fakeStore) Advance(context.Context, int64, Stage, Stage, int) (Campaign, error) {
	return f.item, f.err
}

type fakeModuleLookup struct {
	info ModuleInfo
	err  error
}

func (f *fakeModuleLookup) Latest(context.Context, string) (ModuleInfo, error) {
	return f.info, f.err
}

func TestServiceCreateNormalizesAndAddsChallenge(t *testing.T) {
	store := &fakeStore{item: Campaign{ID: 1, Stage: StageModule}}
	service := NewService(store, nil)

	got, err := service.Create(context.Background(), NewCampaign{
		PlayerName: " Jiayang ",
		Title:      " Go Protocol Run ",
		ModulePath: " github.com/spf13/cobra ",
		Difficulty: DifficultyBeginner,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if store.created.PlayerName != "Jiayang" || store.created.Title != "Go Protocol Run" {
		t.Fatalf("Create input = %#v", store.created)
	}
	if got.CurrentChallenge == nil || got.CurrentChallenge.Stage != StageModule {
		t.Fatalf("Create challenge = %#v", got.CurrentChallenge)
	}
}

func TestServiceGetEnrichesCampaign(t *testing.T) {
	publishedAt := time.Date(2026, time.August, 20, 12, 0, 0, 0, time.UTC)
	service := NewService(
		&fakeStore{item: Campaign{ID: 1, ModulePath: "github.com/spf13/cobra", Stage: StageErrors}},
		&fakeModuleLookup{info: ModuleInfo{LatestVersion: "v1.10.2", PublishedAt: publishedAt}},
	)

	got, err := service.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.ModuleInfo == nil || got.ModuleInfo.LatestVersion != "v1.10.2" {
		t.Fatalf("Get module info = %#v", got.ModuleInfo)
	}
	if got.CurrentChallenge == nil || got.CurrentChallenge.Stage != StageErrors {
		t.Fatalf("Get challenge = %#v", got.CurrentChallenge)
	}
}

func TestServiceListDoesNotRequireProvider(t *testing.T) {
	service := NewService(&fakeStore{items: []Campaign{{ID: 1, Stage: StageChannel}}}, nil)

	got, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(got) != 1 || got[0].CurrentChallenge == nil {
		t.Fatalf("List result = %#v", got)
	}
}

func TestServiceWrapsStoreErrors(t *testing.T) {
	storeErr := errors.New("database unavailable")
	service := NewService(&fakeStore{err: storeErr}, nil)

	_, err := service.Get(context.Background(), 1)
	if !errors.Is(err, storeErr) {
		t.Fatalf("Get error = %v, want wrapped store error", err)
	}
}
