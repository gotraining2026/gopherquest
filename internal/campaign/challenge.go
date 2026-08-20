package campaign

import (
	"fmt"
	"strings"
)

// AttemptOutcome contains the domain result of evaluating an answer.
type AttemptOutcome struct {
	Correct bool
	Message string
	Next    Stage
	Points  int
}

// EvaluateAttempt checks an answer for the current stage and calculates the
// next stage and score delta without persisting anything.
func EvaluateAttempt(stage Stage, answer string, moduleInfo *ModuleInfo, difficulty Difficulty) (AttemptOutcome, error) {
	answer = strings.ToLower(strings.TrimSpace(answer))
	outcome := AttemptOutcome{Next: stage}

	switch stage {
	case StageModule:
		if moduleInfo == nil {
			return AttemptOutcome{}, fmt.Errorf("module challenge requires provider data: %w", ErrProvider)
		}
		major, err := MajorVersionToken(moduleInfo.LatestVersion)
		if err != nil {
			return AttemptOutcome{}, fmt.Errorf("parse module version: %w: %w", ErrProvider, err)
		}
		if answer == strings.ToLower(major) {
			outcome.Correct = true
			outcome.Message = "Correct. The module major version is understood."
			outcome.Next = StageErrors
			outcome.Points = 100 * difficulty.Multiplier()
		} else {
			outcome.Message = "Not quite. Use the major token from latestVersion."
		}
	case StageErrors:
		if answer == "errors.is" {
			outcome.Correct = true
			outcome.Message = "Correct. errors.Is checks the wrapped error chain."
			outcome.Next = StageChannel
			outcome.Points = 100 * difficulty.Multiplier()
		} else {
			outcome.Message = "Not quite. Look for the errors package function that follows wrapping."
		}
	case StageChannel:
		if answer == "channel" {
			outcome.Correct = true
			outcome.Message = "Correct. The campaign is complete."
			outcome.Next = StageComplete
			outcome.Points = 200 * difficulty.Multiplier()
		} else {
			outcome.Message = "Not quite. Think of Go's typed conduit for goroutines."
		}
	default:
		return AttemptOutcome{}, ErrWrongStage
	}

	return outcome, nil
}

// MajorVersionToken returns the vN token from a semantic module version.
func MajorVersionToken(version string) (string, error) {
	version = strings.TrimSpace(version)
	if len(version) < 2 || version[0] != 'v' {
		return "", fmt.Errorf("version %q does not start with v", version)
	}

	dot := strings.IndexByte(version, '.')
	if dot < 0 || dot == 1 {
		return "", fmt.Errorf("version %q has no numeric major token", version)
	}
	for _, char := range version[1:dot] {
		if char < '0' || char > '9' {
			return "", fmt.Errorf("version %q has an invalid major token", version)
		}
	}
	return version[:dot], nil
}
