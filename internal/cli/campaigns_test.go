package cli

import (
	"context"
	"errors"
	"strings"
	"testing"

	"gitlab.galvanize.com/cohorts/ent/sf/sdi-golang/sdi-go-2026-08-10/student-work/jiayang-wang/capstone/internal/api"
)

type fakeClient struct {
	createdRequest  *api.CreateCampaignRequest
	createResponse  api.CreateCampaignRes
	getResponse     api.GetCampaignRes
	attemptRequest  *api.AttemptRequest
	attemptResponse api.SubmitAttemptRes
	err             error
}

func (f *fakeClient) CreateCampaign(_ context.Context, request *api.CreateCampaignRequest) (api.CreateCampaignRes, error) {
	f.createdRequest = request
	return f.createResponse, f.err
}

func (f *fakeClient) GetCampaign(context.Context, api.GetCampaignParams) (api.GetCampaignRes, error) {
	return f.getResponse, f.err
}

func (f *fakeClient) SubmitAttempt(_ context.Context, request *api.AttemptRequest, _ api.SubmitAttemptParams) (api.SubmitAttemptRes, error) {
	f.attemptRequest = request
	return f.attemptResponse, f.err
}

func TestAddCommand(t *testing.T) {
	client := &fakeClient{createResponse: &api.Campaign{ID: 7, Stage: api.StageModule}}
	root := NewRootCommand(client)
	output := new(strings.Builder)
	root.SetOut(output)
	root.SetArgs([]string{"campaigns", "add", "--player", "Jiayang", "--title", "Protocol Run", "--module", "github.com/spf13/cobra", "--difficulty", "advanced"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if client.createdRequest.PlayerName != "Jiayang" || client.createdRequest.Difficulty != api.DifficultyAdvanced {
		t.Fatalf("created request = %#v", client.createdRequest)
	}
	if got := output.String(); got != "Created campaign 7 at stage module\n" {
		t.Fatalf("output = %q", got)
	}
}

func TestGetCommand(t *testing.T) {
	client := &fakeClient{getResponse: &api.Campaign{
		ID:         7,
		PlayerName: "Jiayang",
		Title:      "Protocol Run",
		Stage:      api.StageErrors,
		Score:      100,
	}}
	root := NewRootCommand(client)
	output := new(strings.Builder)
	root.SetOut(output)
	root.SetArgs([]string{"campaigns", "get", "--id", "7"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(output.String(), "Stage: errors") || !strings.Contains(output.String(), "Score: 100") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestAnswerCommand(t *testing.T) {
	client := &fakeClient{attemptResponse: &api.AttemptResult{
		Correct: true,
		Message: "Correct",
		Campaign: api.Campaign{
			ID:    7,
			Title: "Protocol Run",
			Stage: api.StageErrors,
			Score: 100,
		},
	}}
	root := NewRootCommand(client)
	output := new(strings.Builder)
	root.SetOut(output)
	root.SetArgs([]string{"campaigns", "answer", "--id", "7", "--answer", "v1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if client.attemptRequest.Answer != "v1" || !strings.Contains(output.String(), "Correct") {
		t.Fatalf("request/output = %#v / %q", client.attemptRequest, output.String())
	}
}

func TestRequiredFlagsAndReturnedErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "add player", args: []string{"campaigns", "add", "--title", "Title", "--module", "github.com/spf13/cobra"}},
		{name: "get id", args: []string{"campaigns", "get"}},
		{name: "answer", args: []string{"campaigns", "answer", "--id", "1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := NewRootCommand(&fakeClient{})
			root.SetArgs(tt.args)
			if err := root.Execute(); err == nil {
				t.Fatal("Execute returned nil error")
			}
		})
	}

	clientErr := errors.New("api unavailable")
	root := NewRootCommand(&fakeClient{err: clientErr})
	root.SetArgs([]string{"campaigns", "get", "--id", "1"})
	if err := root.Execute(); !errors.Is(err, clientErr) {
		t.Fatalf("Execute error = %v, want wrapped client error", err)
	}
}
