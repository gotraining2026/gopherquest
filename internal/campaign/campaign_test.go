package campaign

import (
	"errors"
	"testing"
)

func TestNewCampaignValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   NewCampaign
		wantErr bool
	}{
		{
			name: "valid",
			input: NewCampaign{
				PlayerName: "Jiayang",
				Title:      "Go Protocol Run",
				ModulePath: "github.com/spf13/cobra",
				Difficulty: DifficultyBeginner,
			},
		},
		{
			name: "trims are rejected by domain validation",
			input: NewCampaign{
				PlayerName: " Jiayang",
				Title:      "Go Protocol Run",
				ModulePath: "github.com/spf13/cobra",
				Difficulty: DifficultyBeginner,
			},
			wantErr: true,
		},
		{
			name: "invalid module path",
			input: NewCampaign{
				PlayerName: "Jiayang",
				Title:      "Go Protocol Run",
				ModulePath: "github.com/Spf13/cobra",
				Difficulty: DifficultyBeginner,
			},
			wantErr: true,
		},
		{
			name: "valid major suffix",
			input: NewCampaign{
				PlayerName: "Jiayang",
				Title:      "Go Protocol Run",
				ModulePath: "github.com/spf13/cobra/v2",
				Difficulty: DifficultyAdvanced,
			},
		},
		{
			name: "invalid difficulty",
			input: NewCampaign{
				PlayerName: "Jiayang",
				Title:      "Go Protocol Run",
				ModulePath: "github.com/spf13/cobra",
				Difficulty: Difficulty("expert"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, ErrInvalid) {
				t.Fatalf("Validate() error = %v, want ErrInvalid", err)
			}
		})
	}
}

func TestDifficultyMultiplier(t *testing.T) {
	tests := []struct {
		difficulty Difficulty
		want       int
	}{
		{difficulty: DifficultyBeginner, want: 1},
		{difficulty: DifficultyIntermediate, want: 2},
		{difficulty: DifficultyAdvanced, want: 3},
		{difficulty: Difficulty("unknown"), want: 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.difficulty), func(t *testing.T) {
			if got := tt.difficulty.Multiplier(); got != tt.want {
				t.Fatalf("Multiplier() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestChallengeFor(t *testing.T) {
	if got := ChallengeFor(StageComplete); got != nil {
		t.Fatalf("ChallengeFor(complete) = %#v, want nil", got)
	}

	for _, stage := range []Stage{StageModule, StageErrors, StageChannel} {
		t.Run(string(stage), func(t *testing.T) {
			challenge := ChallengeFor(stage)
			if challenge == nil {
				t.Fatal("ChallengeFor() returned nil")
			}
			if challenge.Stage != stage {
				t.Fatalf("ChallengeFor().Stage = %q, want %q", challenge.Stage, stage)
			}
			if challenge.Prompt == "" || challenge.Hint == "" {
				t.Fatal("ChallengeFor() returned an incomplete challenge")
			}
		})
	}
}
