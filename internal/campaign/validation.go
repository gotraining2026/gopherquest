package campaign

import (
	"fmt"
	"regexp"
	"strings"
)

var modulePathPattern = regexp.MustCompile(`^github\.com/[a-z0-9][a-z0-9.-]*/[a-z0-9][a-z0-9.-]*(/v[2-9][0-9]*)?$`)

// Validate checks the client-editable fields for a new campaign.
func (c NewCampaign) Validate() error {
	if err := validateProfile(c.PlayerName, c.Title, c.ModulePath, c.Difficulty); err != nil {
		return fmt.Errorf("validate new campaign: %w", err)
	}
	return nil
}

// Validate checks the client-editable fields for a campaign replacement.
func (c UpdateCampaign) Validate() error {
	if err := validateProfile(c.PlayerName, c.Title, c.ModulePath, c.Difficulty); err != nil {
		return fmt.Errorf("validate campaign update: %w", err)
	}
	return nil
}

func validateProfile(playerName, title, modulePath string, difficulty Difficulty) error {
	if playerName != strings.TrimSpace(playerName) || len(playerName) < 1 || len(playerName) > 40 {
		return fmt.Errorf("playerName must be trimmed and 1-40 characters: %w", ErrInvalid)
	}
	if title != strings.TrimSpace(title) || len(title) < 1 || len(title) > 80 {
		return fmt.Errorf("title must be trimmed and 1-80 characters: %w", ErrInvalid)
	}
	if len(modulePath) < 3 || len(modulePath) > 200 || !modulePathPattern.MatchString(modulePath) {
		return fmt.Errorf("modulePath must be a lower-case GitHub module path: %w", ErrInvalid)
	}
	if difficulty.Multiplier() == 0 {
		return fmt.Errorf("difficulty is not supported: %w", ErrInvalid)
	}
	return nil
}
