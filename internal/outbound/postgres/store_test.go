package repository

import (
	"testing"
	"time"

	"gitlab.galvanize.com/cohorts/ent/sf/sdi-golang/sdi-go-2026-08-10/student-work/jiayang-wang/capstone/internal/campaign"
)

func TestToCampaign(t *testing.T) {
	createdAt := time.Date(2026, time.August, 20, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)

	got := toCampaign(Campaign{
		ID:         7,
		PlayerName: "Jiayang",
		Title:      "Go Protocol Run",
		ModulePath: "github.com/spf13/cobra",
		Difficulty: "intermediate",
		Stage:      "errors",
		Score:      200,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	})

	if got.ID != 7 || got.Difficulty != campaign.DifficultyIntermediate || got.Stage != campaign.StageErrors {
		t.Fatalf("toCampaign() = %#v", got)
	}
	if !got.CreatedAt.Equal(createdAt) || !got.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("toCampaign() timestamps = %v, %v", got.CreatedAt, got.UpdatedAt)
	}
}
