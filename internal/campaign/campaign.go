package campaign

import (
	"context"
	"time"
)

// Difficulty controls the score multiplier for a campaign.
type Difficulty string

const (
	// DifficultyBeginner is the one-times score level.
	DifficultyBeginner Difficulty = "beginner"
	// DifficultyIntermediate is the two-times score level.
	DifficultyIntermediate Difficulty = "intermediate"
	// DifficultyAdvanced is the three-times score level.
	DifficultyAdvanced Difficulty = "advanced"
)

// Stage identifies the current challenge in a campaign.
type Stage string

const (
	// StageModule is the module-version challenge.
	StageModule Stage = "module"
	// StageErrors is the wrapped-error challenge.
	StageErrors Stage = "errors"
	// StageChannel is the channel vocabulary challenge.
	StageChannel Stage = "channel"
	// StageComplete identifies a completed campaign.
	StageComplete Stage = "complete"
)

// Campaign is the persisted game resource plus response-only enrichment.
type Campaign struct {
	ID               int64
	PlayerName       string
	Title            string
	ModulePath       string
	Difficulty       Difficulty
	Stage            Stage
	Score            int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ModuleInfo       *ModuleInfo
	CurrentChallenge *Challenge
}

// NewCampaign contains the client-editable values used at creation time.
type NewCampaign struct {
	PlayerName string
	Title      string
	ModulePath string
	Difficulty Difficulty
}

// UpdateCampaign contains the client-editable values used for replacement.
type UpdateCampaign struct {
	PlayerName string
	Title      string
	ModulePath string
	Difficulty Difficulty
}

// ModuleInfo contains live metadata returned by the external module provider.
type ModuleInfo struct {
	LatestVersion string
	PublishedAt   time.Time
}

// Challenge contains the player-facing prompt for a stage.
type Challenge struct {
	Stage  Stage
	Prompt string
	Hint   string
}

// AttemptResult contains the answer result and the resulting campaign state.
type AttemptResult struct {
	Correct  bool
	Message  string
	Campaign Campaign
}

// Store persists campaigns and performs atomic stage advancement.
type Store interface {
	Create(context.Context, NewCampaign) (Campaign, error)
	List(context.Context) ([]Campaign, error)
	Get(context.Context, int64) (Campaign, error)
	Update(context.Context, int64, UpdateCampaign) (Campaign, error)
	Delete(context.Context, int64) error
	Advance(context.Context, int64, Stage, Stage, int) (Campaign, error)
}

// ModuleLookup retrieves live module metadata for a campaign.
type ModuleLookup interface {
	Latest(context.Context, string) (ModuleInfo, error)
}

// Multiplier returns the score multiplier for a difficulty.
func (d Difficulty) Multiplier() int {
	switch d {
	case DifficultyBeginner:
		return 1
	case DifficultyIntermediate:
		return 2
	case DifficultyAdvanced:
		return 3
	default:
		return 0
	}
}

// ChallengeFor returns the stable challenge metadata for a stage.
func ChallengeFor(stage Stage) *Challenge {
	switch stage {
	case StageModule:
		return &Challenge{
			Stage:  StageModule,
			Prompt: "What semantic-version major token starts the latest release?",
			Hint:   "Look at the latestVersion returned for the module.",
		}
	case StageErrors:
		return &Challenge{
			Stage:  StageErrors,
			Prompt: "Which standard-library function checks a wrapped ErrNotFound chain?",
			Hint:   "The answer begins with errors.",
		}
	case StageChannel:
		return &Challenge{
			Stage:  StageChannel,
			Prompt: "Which typed conduit lets goroutines synchronize by communicating?",
			Hint:   "It is a core Go concurrency vocabulary word.",
		}
	default:
		return nil
	}
}
