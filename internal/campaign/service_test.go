package campaign

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeStore struct {
	created      NewCampaign
	updated      UpdateCampaign
	item         Campaign
	items        []Campaign
	err          error
	createCalls  int
	advanceCalls int
}

func (f *fakeStore) Create(_ context.Context, input NewCampaign) (Campaign, error) {
	f.createCalls++
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

func (f *fakeStore) Update(_ context.Context, _ int64, input UpdateCampaign) (Campaign, error) {
	f.updated = input
	if f.err != nil {
		return Campaign{}, f.err
	}
	return f.item, nil
}

func (f *fakeStore) Delete(context.Context, int64) error {
	return f.err
}

func (f *fakeStore) Advance(_ context.Context, _ int64, expected, next Stage, points int) (Campaign, error) {
	f.advanceCalls++
	if f.err != nil {
		return Campaign{}, f.err
	}
	if f.item.Stage != expected {
		return Campaign{}, ErrWrongStage
	}
	f.item.Stage = next
	f.item.Score += points
	return f.item, nil
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

func TestServiceCreateRejectsInvalidInputBeforeStore(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, nil)

	_, err := service.Create(context.Background(), NewCampaign{
		PlayerName: "player",
		Title:      "title",
		ModulePath: "not-a-module",
		Difficulty: DifficultyBeginner,
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Create error = %v, want ErrInvalid", err)
	}
	if store.createCalls != 0 {
		t.Fatalf("Create called store %d times, want 0", store.createCalls)
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

func TestServiceUpdateNormalizesInput(t *testing.T) {
	store := &fakeStore{item: Campaign{ID: 1, Stage: StageModule}}
	service := NewService(store, nil)

	_, err := service.Update(context.Background(), 1, UpdateCampaign{
		PlayerName: " player ",
		Title:      " title ",
		ModulePath: " github.com/spf13/cobra ",
		Difficulty: DifficultyIntermediate,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if store.updated.PlayerName != "player" || store.updated.Title != "title" || store.updated.ModulePath != "github.com/spf13/cobra" {
		t.Fatalf("Update input = %#v", store.updated)
	}
}

func TestServiceGetWrapsProviderError(t *testing.T) {
	service := NewService(
		&fakeStore{item: Campaign{ID: 1, ModulePath: "github.com/spf13/cobra"}},
		&fakeModuleLookup{err: ErrProvider},
	)

	_, err := service.Get(context.Background(), 1)
	if !errors.Is(err, ErrProvider) {
		t.Fatalf("Get error = %v, want ErrProvider", err)
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

func TestServiceAttemptWrongAnswerDoesNotAdvance(t *testing.T) {
	store := &fakeStore{item: Campaign{
		ID:         1,
		ModulePath: "github.com/spf13/cobra",
		Difficulty: DifficultyBeginner,
		Stage:      StageModule,
	}}
	service := NewService(store, &fakeModuleLookup{info: ModuleInfo{LatestVersion: "v1.10.2"}})

	result, err := service.Attempt(context.Background(), 1, "v2")
	if err != nil {
		t.Fatalf("Attempt returned error: %v", err)
	}
	if result.Correct || store.advanceCalls != 0 || result.Campaign.Stage != StageModule {
		t.Fatalf("wrong attempt result = %#v, advanceCalls = %d", result, store.advanceCalls)
	}
}

func TestServiceAttemptCorrectAdvances(t *testing.T) {
	store := &fakeStore{item: Campaign{
		ID:         1,
		ModulePath: "github.com/spf13/cobra",
		Difficulty: DifficultyIntermediate,
		Stage:      StageModule,
	}}
	service := NewService(store, &fakeModuleLookup{info: ModuleInfo{LatestVersion: "v1.10.2"}})

	result, err := service.Attempt(context.Background(), 1, "v1")
	if err != nil {
		t.Fatalf("Attempt returned error: %v", err)
	}
	if !result.Correct || store.advanceCalls != 1 {
		t.Fatalf("correct attempt result = %#v, advanceCalls = %d", result, store.advanceCalls)
	}
	if result.Campaign.Stage != StageErrors || result.Campaign.Score != 200 {
		t.Fatalf("service result = %#v, want errors stage and 200 points", result.Campaign)
	}
}
