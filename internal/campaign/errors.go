package campaign

import "errors"

var (
	// ErrNotFound indicates that a campaign does not exist.
	ErrNotFound = errors.New("campaign not found")
	// ErrInvalid indicates that a campaign value is invalid.
	ErrInvalid = errors.New("invalid campaign")
	// ErrWrongStage indicates that an attempt does not match current progress.
	ErrWrongStage = errors.New("attempt does not match current stage")
)
