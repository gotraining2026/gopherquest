package campaign

import "testing"

func TestMajorVersionToken(t *testing.T) {
	tests := []struct {
		version string
		want    string
		wantErr bool
	}{
		{version: "v1.10.2", want: "v1"},
		{version: "v20.0.0", want: "v20"},
		{version: "v1", wantErr: true},
		{version: "1.10.2", wantErr: true},
		{version: "vX.10.2", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got, err := MajorVersionToken(tt.version)
			if (err != nil) != tt.wantErr {
				t.Fatalf("MajorVersionToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("MajorVersionToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEvaluateAttempt(t *testing.T) {
	moduleInfo := &ModuleInfo{LatestVersion: "v1.10.2"}
	tests := []struct {
		name        string
		stage       Stage
		answer      string
		wantNext    Stage
		wantPoints  int
		wantCorrect bool
	}{
		{name: "module correct", stage: StageModule, answer: " V1 ", wantNext: StageErrors, wantPoints: 100, wantCorrect: true},
		{name: "module wrong", stage: StageModule, answer: "v2", wantNext: StageModule, wantCorrect: false},
		{name: "errors normalized", stage: StageErrors, answer: " Errors.Is ", wantNext: StageChannel, wantPoints: 100, wantCorrect: true},
		{name: "channel correct", stage: StageChannel, answer: "channel", wantNext: StageComplete, wantPoints: 200, wantCorrect: true},
		{name: "channel wrong", stage: StageChannel, answer: "goroutine", wantNext: StageChannel, wantCorrect: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := moduleInfo
			if tt.stage != StageModule {
				info = nil
			}
			got, err := EvaluateAttempt(tt.stage, tt.answer, info, DifficultyBeginner)
			if err != nil {
				t.Fatalf("EvaluateAttempt returned error: %v", err)
			}
			if got.Next != tt.wantNext || got.Points != tt.wantPoints || got.Correct != tt.wantCorrect {
				t.Fatalf("EvaluateAttempt() = %#v", got)
			}
		})
	}
}
