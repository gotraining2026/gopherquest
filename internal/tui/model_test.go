package tui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"gitlab.galvanize.com/cohorts/ent/sf/sdi-golang/sdi-go-2026-08-10/student-work/jiayang-wang/capstone/internal/api"
)

type fakeClient struct {
	createRequest    *api.CreateCampaignRequest
	createResponse   api.CreateCampaignRes
	deleteID         int64
	deleteResponse   api.DeleteCampaignRes
	campaignResponse api.GetCampaignRes
	listResponse     api.ListCampaignsRes
	attemptResponse  api.SubmitAttemptRes
	attemptRequest   *api.AttemptRequest
}

func (f *fakeClient) CreateCampaign(_ context.Context, request *api.CreateCampaignRequest) (api.CreateCampaignRes, error) {
	f.createRequest = request
	return f.createResponse, nil
}

func (f *fakeClient) DeleteCampaign(_ context.Context, params api.DeleteCampaignParams) (api.DeleteCampaignRes, error) {
	f.deleteID = params.ID
	return f.deleteResponse, nil
}

func (f *fakeClient) GetCampaign(context.Context, api.GetCampaignParams) (api.GetCampaignRes, error) {
	return f.campaignResponse, nil
}

func (f *fakeClient) ListCampaigns(context.Context) (api.ListCampaignsRes, error) {
	return f.listResponse, nil
}

func (f *fakeClient) SubmitAttempt(_ context.Context, request *api.AttemptRequest, _ api.SubmitAttemptParams) (api.SubmitAttemptRes, error) {
	f.attemptRequest = request
	return f.attemptResponse, nil
}

func TestModelLoadsAndRendersCampaign(t *testing.T) {
	campaign := &api.Campaign{
		ID:    1,
		Stage: api.StageModule,
		Score: 0,
	}
	m := New(&fakeClient{campaignResponse: campaign}, 1).(model)

	updated, cmd := m.Update(campaignLoadedMsg{campaign: campaign})
	if cmd == nil {
		t.Fatal("Update did not start the animation command")
	}
	view := updated.(model).View()
	if !strings.Contains(view.Content, "______") || !strings.Contains(view.Content, "Stage: module") {
		t.Fatalf("view = %q", view.Content)
	}
}

func TestModelSubmitsAnswerAsCommand(t *testing.T) {
	client := &fakeClient{
		campaignResponse: &api.Campaign{ID: 1, Stage: api.StageModule},
		attemptResponse: &api.AttemptResult{
			Correct: true,
			Campaign: api.Campaign{
				ID:    1,
				Stage: api.StageErrors,
				Score: 100,
			},
		},
	}
	m := New(client, 1).(model)
	updated, _ := m.Update(campaignLoadedMsg{campaign: client.campaignResponse.(*api.Campaign)})
	m = updated.(model)

	updated, cmd := m.Update(tea.KeyPressMsg{Text: "v1", Code: 'v'})
	m = updated.(model)
	if m.answer != "v1" {
		t.Fatalf("answer = %q, want v1", m.answer)
	}
	updated, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(model)
	if !m.loading || cmd == nil {
		t.Fatalf("model/cmd after enter = %#v / %v", m, cmd)
	}
	msg := cmd()
	if _, ok := msg.(errorMsg); ok {
		t.Fatalf("command returned error message: %#v", msg)
	}
	if client.attemptRequest == nil || client.attemptRequest.Answer != "v1" {
		t.Fatalf("attempt request = %#v", client.attemptRequest)
	}
}

func TestModelRendersCompletion(t *testing.T) {
	m := New(&fakeClient{}, 1).(model)
	m.returnToList = true
	completed := &api.AttemptResult{
		Correct: true,
		Message: "Correct. The campaign is complete.",
		Campaign: api.Campaign{
			ID:    1,
			Stage: api.StageComplete,
			Score: 400,
		},
	}
	updated, cmd := m.Update(attemptLoadedMsg{result: completed})
	if cmd != nil {
		t.Fatal("Update returned an unexpected command")
	}
	view := updated.(model).View()
	if !strings.Contains(view.Content, "FINISH") || !strings.Contains(view.Content, "Score: 400") {
		t.Fatalf("view = %q", view.Content)
	}
	if strings.Contains(view.Content, "Enter: submit") || strings.Contains(view.Content, "Backspace: delete") {
		t.Fatalf("completed view has inactive commands: %q", view.Content)
	}
	if !strings.Contains(view.Content, "Esc: campaigns") || !strings.Contains(view.Content, "Q: quit") {
		t.Fatalf("completed view is missing navigation commands: %q", view.Content)
	}
}

func TestModelListsAndSelectsCampaign(t *testing.T) {
	items := api.ListCampaignsOKApplicationJSON{
		{ID: 1, Title: "First Run", Stage: api.StageModule},
		{ID: 2, Title: "Second Run", Stage: api.StageErrors},
	}
	client := &fakeClient{listResponse: &items, campaignResponse: &api.Campaign{ID: 2, Stage: api.StageErrors}}
	m := New(client, 0).(model)

	updated, cmd := m.Update(campaignsLoadedMsg{campaigns: []api.Campaign(items)})
	if cmd != nil {
		t.Fatal("list update returned an unexpected command")
	}
	m = updated.(model)
	if !strings.Contains(m.View().Content, "First Run") {
		t.Fatalf("list view = %q", m.View().Content)
	}

	updated, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(model)
	if !m.loading || cmd == nil {
		t.Fatalf("select result = %#v / %v", m, cmd)
	}
	if client.campaignResponse == nil {
		t.Fatal("selection did not retain a campaign response")
	}
}

func TestModelCreatesCampaignFromForm(t *testing.T) {
	client := &fakeClient{
		createResponse:   &api.Campaign{ID: 9, Stage: api.StageModule},
		campaignResponse: &api.Campaign{ID: 9, Stage: api.StageModule, ModulePath: "github.com/spf13/cobra"},
	}
	m := New(client, 0).(model)
	m.screen = newCampaignScreen
	m.formValues = [4]string{"Jiayang", "New Run", "github.com/spf13/cobra", "beginner"}

	msg := m.createCampaignCmd()()
	created, ok := msg.(campaignCreatedMsg)
	if !ok || created.campaign.ID != 9 {
		t.Fatalf("create message = %#v", msg)
	}
	if client.createRequest.Title != "New Run" {
		t.Fatalf("create request = %#v", client.createRequest)
	}
	updated, cmd := m.Update(msg)
	if cmd == nil || !updated.(model).loading || updated.(model).screen != gameScreen {
		t.Fatalf("created model = %#v / %v", updated, cmd)
	}
	loaded, ok := cmd().(campaignLoadedMsg)
	if !ok || loaded.campaign.ID != 9 {
		t.Fatalf("enrichment message = %#v", loaded)
	}
}

func TestNewCampaignFormShowsTitlePlaceholder(t *testing.T) {
	m := New(&fakeClient{}, 0).(model)
	m.loading = false
	updated, _ := m.Update(tea.KeyPressMsg{Text: "n", Code: 'n'})
	m = updated.(model)

	got := m.View().Content
	if !strings.Contains(got, "(e.g., Go Protocol Run)") {
		t.Fatalf("new campaign view = %q", got)
	}
	if !strings.Contains(got, "github.com/spf13/cobra") {
		t.Fatalf("new campaign view does not show default module preset: %q", got)
	}
}

func TestModulePresetCycle(t *testing.T) {
	if got := cycleModulePreset(modulePresets[0], 1); got != modulePresets[1] {
		t.Fatalf("cycle forward = %q, want %q", got, modulePresets[1])
	}
	if got := cycleModulePreset(modulePresets[0], -1); got != modulePresets[len(modulePresets)-1] {
		t.Fatalf("cycle backward = %q, want %q", got, modulePresets[len(modulePresets)-1])
	}
}

func TestDifficultyPresetCycle(t *testing.T) {
	if got := cyclePreset("beginner", difficultyPresets, 1); got != "intermediate" {
		t.Fatalf("difficulty cycle forward = %q, want intermediate", got)
	}
	if got := cyclePreset("beginner", difficultyPresets, -1); got != "advanced" {
		t.Fatalf("difficulty cycle backward = %q, want advanced", got)
	}
}

func TestSceneProgressesAcrossStages(t *testing.T) {
	tests := []struct {
		stage api.Stage
		pits  int
	}{
		{stage: api.StageModule, pits: 3},
		{stage: api.StageErrors, pits: 2},
		{stage: api.StageChannel, pits: 1},
		{stage: api.StageComplete, pits: 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.stage), func(t *testing.T) {
			got := sceneForStage(tt.stage, false)
			if strings.Count(got, "______") != tt.pits {
				t.Fatalf("scene = %q, pits = %d, want %d", got, strings.Count(got, "______"), tt.pits)
			}
		})
	}
	if sceneForStage(api.StageModule, false) == sceneForStage(api.StageModule, true) {
		t.Fatal("running frames should have different gopher poses")
	}
}

func TestModelDeletesSelectedCampaign(t *testing.T) {
	items := []api.Campaign{{ID: 4, Title: "Delete Me", Stage: api.StageModule}}
	client := &fakeClient{deleteResponse: &api.DeleteCampaignNoContent{}}
	m := New(client, 0).(model)
	m.screen = campaignListScreen
	m.campaigns = items
	m.loading = false

	updated, cmd := m.Update(tea.KeyPressMsg{Text: "d", Code: 'd'})
	m = updated.(model)
	if !m.loading || cmd == nil {
		t.Fatalf("delete result = %#v / %v", m, cmd)
	}
	updated, cmd = m.Update(cmd())
	m = updated.(model)
	if cmd != nil || client.deleteID != 4 || len(m.campaigns) != 0 {
		t.Fatalf("deleted model = %#v / %v", m, cmd)
	}
}

func TestRemoveLastRune(t *testing.T) {
	if got := removeLastRune("go"); got != "g" {
		t.Fatalf("removeLastRune(go) = %q", got)
	}
	if got := removeLastRune("界"); got != "" {
		t.Fatalf("removeLastRune(界) = %q", got)
	}
}
