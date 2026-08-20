//go:build integration

package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"gitlab.galvanize.com/cohorts/ent/sf/sdi-golang/sdi-go-2026-08-10/student-work/jiayang-wang/capstone/internal/campaign"
)

func TestStoreIntegration(t *testing.T) {
	dsn := os.Getenv("GOPHERQUEST_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("GOPHERQUEST_TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	defer db.Close()

	if got := db.Stats().MaxOpenConnections; got != maxOpenConnections {
		t.Fatalf("MaxOpenConnections = %d, want %d", got, maxOpenConnections)
	}

	store := NewStore(db)
	created, err := store.Create(ctx, campaign.NewCampaign{
		PlayerName: "Integration Player",
		Title:      "Persistence Run",
		ModulePath: "github.com/spf13/cobra",
		Difficulty: campaign.DifficultyIntermediate,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Delete(context.Background(), created.ID)
	})

	got, err := store.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.Title != "Persistence Run" || got.Score != 0 || got.Stage != campaign.StageModule {
		t.Fatalf("Get returned campaign: %#v", got)
	}

	updated, err := store.Update(ctx, created.ID, campaign.UpdateCampaign{
		PlayerName: "Updated Player",
		Title:      "Updated Run",
		ModulePath: "github.com/spf13/cobra",
		Difficulty: campaign.DifficultyAdvanced,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.PlayerName != "Updated Player" || updated.Difficulty != campaign.DifficultyAdvanced {
		t.Fatalf("Update returned campaign: %#v", updated)
	}

	advanced, err := store.Advance(ctx, created.ID, campaign.StageModule, campaign.StageErrors, 200)
	if err != nil {
		t.Fatalf("Advance returned error: %v", err)
	}
	if advanced.Stage != campaign.StageErrors || advanced.Score != 200 {
		t.Fatalf("Advance returned campaign: %#v", advanced)
	}

	if _, err := store.Update(ctx, created.ID, campaign.UpdateCampaign{
		PlayerName: "Too Late",
		Title:      "Too Late",
		ModulePath: "github.com/spf13/cobra",
		Difficulty: campaign.DifficultyBeginner,
	}); !errors.Is(err, campaign.ErrWrongStage) {
		t.Fatalf("Update after advancement error = %v, want ErrWrongStage", err)
	}

	if _, err := store.Advance(ctx, created.ID, campaign.StageModule, campaign.StageChannel, 100); !errors.Is(err, campaign.ErrWrongStage) {
		t.Fatalf("wrong-stage Advance error = %v, want ErrWrongStage", err)
	}

	if err := store.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if _, err := store.Get(ctx, created.ID); !errors.Is(err, campaign.ErrNotFound) {
		t.Fatalf("Get after delete error = %v, want ErrNotFound", err)
	}
}
