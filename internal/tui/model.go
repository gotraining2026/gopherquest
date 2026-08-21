package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"gitlab.galvanize.com/cohorts/ent/sf/sdi-golang/sdi-go-2026-08-10/student-work/jiayang-wang/capstone/internal/api"
)

// Client is the generated API surface needed by the campaign TUI.
type Client interface {
	CreateCampaign(context.Context, *api.CreateCampaignRequest) (api.CreateCampaignRes, error)
	DeleteCampaign(context.Context, api.DeleteCampaignParams) (api.DeleteCampaignRes, error)
	GetCampaign(context.Context, api.GetCampaignParams) (api.GetCampaignRes, error)
	ListCampaigns(context.Context) (api.ListCampaignsRes, error)
	SubmitAttempt(context.Context, *api.AttemptRequest, api.SubmitAttemptParams) (api.SubmitAttemptRes, error)
}

type screen uint8

const (
	campaignListScreen screen = iota
	newCampaignScreen
	gameScreen
)

const (
	playerField = iota
	titleField
	moduleField
	difficultyField
)

var formLabels = [...]string{"Player", "Title", "Module", "Difficulty"}

var modulePresets = []string{
	"github.com/spf13/cobra",
	"github.com/ogen-go/ogen",
	"github.com/charmbracelet/bubbletea/v2",
}

var difficultyPresets = []string{"beginner", "intermediate", "advanced"}

type model struct {
	client       Client
	screen       screen
	campaignID   int64
	campaign     *api.Campaign
	campaigns    []api.Campaign
	selected     int
	returnToList bool
	answer       string
	status       string
	statusGood   bool
	err          error
	loading      bool
	legFrame     bool
	formField    int
	formValues   [4]string
}

var (
	titleStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	menuStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	selectedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	stageStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
	promptStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	hintStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	placeholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	successStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	gopherStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	groundStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	pitStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	finishStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
)

type campaignLoadedMsg struct {
	campaign *api.Campaign
}

type campaignsLoadedMsg struct {
	campaigns []api.Campaign
}

type campaignCreatedMsg struct {
	campaign *api.Campaign
}

type campaignDeletedMsg struct {
	id int64
}

type attemptLoadedMsg struct {
	result *api.AttemptResult
}

type errorMsg struct {
	err error
}

type animationTickMsg struct{}

// New builds a campaign manager or a direct game screen when campaignID is set.
func New(client Client, campaignID int64) tea.Model {
	m := model{client: client, campaignID: campaignID, loading: true}
	if campaignID > 0 {
		m.screen = gameScreen
		m.returnToList = false
	}
	return m
}

// Run starts the campaign TUI and returns any terminal error.
func Run(client Client, campaignID int64) error {
	program := tea.NewProgram(New(client, campaignID))
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("run campaign TUI: %w", err)
	}
	return nil
}

func (m model) Init() tea.Cmd {
	if m.campaignID > 0 {
		return tea.Batch(m.getCampaignCmd(), animationTickCmd())
	}
	return m.listCampaignsCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case campaignLoadedMsg:
		m.campaign = msg.campaign
		m.campaignID = msg.campaign.ID
		m.screen = gameScreen
		m.loading = false
		m.err = nil
		return m, animationTickCmd()
	case campaignsLoadedMsg:
		m.campaigns = msg.campaigns
		m.screen = campaignListScreen
		m.loading = false
		m.err = nil
		m.selected = clampSelection(m.selected, len(m.campaigns))
		return m, nil
	case campaignCreatedMsg:
		m.campaign = msg.campaign
		m.campaignID = msg.campaign.ID
		m.screen = gameScreen
		m.returnToList = true
		m.status = fmt.Sprintf("Created campaign %d.", msg.campaign.ID)
		m.statusGood = true
		m.loading = true
		m.err = nil
		return m, m.getCampaignCmd()
	case campaignDeletedMsg:
		m.campaigns = removeCampaign(m.campaigns, msg.id)
		m.selected = clampSelection(m.selected, len(m.campaigns))
		m.status = fmt.Sprintf("Deleted campaign %d.", msg.id)
		m.statusGood = true
		m.loading = false
		m.err = nil
		return m, nil
	case attemptLoadedMsg:
		m.campaign = &msg.result.Campaign
		m.answer = ""
		m.status = msg.result.Message
		m.statusGood = msg.result.Correct
		m.loading = false
		m.err = nil
		return m, nil
	case errorMsg:
		m.err = msg.err
		m.loading = false
		return m, nil
	case animationTickMsg:
		if m.screen != gameScreen || m.campaign == nil || m.campaign.Stage == api.StageComplete {
			return m, nil
		}
		m.legFrame = !m.legFrame
		return m, animationTickCmd()
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	}
	return m, nil
}

func (m model) updateKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.screen {
	case campaignListScreen:
		return m.updateListKey(msg)
	case newCampaignScreen:
		return m.updateNewKey(msg)
	case gameScreen:
		return m.updateGameKey(msg)
	default:
		return m, nil
	}
}

func (m model) updateListKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "up", "k":
		m.selected = clampSelection(m.selected-1, len(m.campaigns))
	case "down", "j":
		m.selected = clampSelection(m.selected+1, len(m.campaigns))
	case "enter":
		if len(m.campaigns) == 0 {
			return m, nil
		}
		m.campaignID = m.campaigns[m.selected].ID
		m.campaign = nil
		m.returnToList = true
		m.screen = gameScreen
		m.loading = true
		return m, m.getCampaignCmd()
	case "n":
		m.formValues = [4]string{"", "", modulePresets[0], "beginner"}
		m.formField = playerField
		m.status = ""
		m.err = nil
		m.screen = newCampaignScreen
	case "d":
		if len(m.campaigns) == 0 || m.loading {
			return m, nil
		}
		m.loading = true
		return m, m.deleteCampaignCmd(m.campaigns[m.selected].ID)
	case "r":
		m.loading = true
		return m, m.listCampaignsCmd()
	}
	return m, nil
}

func (m model) updateNewKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = campaignListScreen
		m.err = nil
		return m, m.listCampaignsCmd()
	case "backspace":
		m.formValues[m.formField] = removeLastRune(m.formValues[m.formField])
	case "tab", "down":
		m.formField = (m.formField + 1) % len(formLabels)
	case "shift+tab", "up":
		m.formField = (m.formField + len(formLabels) - 1) % len(formLabels)
	case "left", "h":
		if m.formField == moduleField {
			m.formValues[moduleField] = cycleModulePreset(m.formValues[moduleField], -1)
		} else if m.formField == difficultyField {
			m.formValues[difficultyField] = cyclePreset(m.formValues[difficultyField], difficultyPresets, -1)
		}
	case "right", "l":
		if m.formField == moduleField {
			m.formValues[moduleField] = cycleModulePreset(m.formValues[moduleField], 1)
		} else if m.formField == difficultyField {
			m.formValues[difficultyField] = cyclePreset(m.formValues[difficultyField], difficultyPresets, 1)
		}
	case "enter":
		if m.formField < difficultyField {
			m.formField++
			return m, nil
		}
		m.loading = true
		m.err = nil
		return m, m.createCampaignCmd()
	default:
		if text := msg.Key().Text; text != "" {
			if m.formField == difficultyField {
				return m, nil
			}
			if m.formField == moduleField && isModulePreset(m.formValues[moduleField]) {
				m.formValues[moduleField] = ""
			}
			m.formValues[m.formField] += text
			m.err = nil
		}
	}
	return m, nil
}

func (m model) updateGameKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "esc":
		if m.returnToList {
			m.screen = campaignListScreen
			m.campaign = nil
			m.status = ""
			m.err = nil
			return m, m.listCampaignsCmd()
		}
		return m, tea.Quit
	case "backspace":
		m.answer = removeLastRune(m.answer)
	case "enter":
		if m.loading || m.campaign == nil || m.answer == "" || m.campaign.Stage == api.StageComplete {
			return m, nil
		}
		m.loading = true
		return m, m.submitAttemptCmd(m.answer)
	default:
		if text := msg.Key().Text; text != "" {
			m.answer += text
			m.err = nil
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	switch m.screen {
	case campaignListScreen:
		return m.viewCampaignList()
	case newCampaignScreen:
		return m.viewNewCampaign()
	case gameScreen:
		return m.viewGame()
	default:
		return newView("GopherQuest\n")
	}
}

func (m model) viewCampaignList() tea.View {
	var body strings.Builder
	body.WriteString(titleStyle.Render("GopherQuest"))
	body.WriteString("\n\n")
	if m.loading {
		body.WriteString("Loading campaigns...\n")
		return newView(body.String())
	}
	if m.err != nil {
		body.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		body.WriteString("\n\nPress q to quit.\n")
		return newView(body.String())
	}
	body.WriteString(menuStyle.Render("Campaigns"))
	body.WriteString("\n\n")
	if len(m.campaigns) == 0 {
		body.WriteString("No campaigns found.\n")
	} else {
		for i, item := range m.campaigns {
			line := fmt.Sprintf("  %d  %-24s  %-8s  score=%d", item.ID, item.Title, item.Stage, item.Score)
			if i == m.selected {
				line = ">" + line[1:]
				body.WriteString(selectedStyle.Render(line))
			} else {
				body.WriteString(line)
			}
			body.WriteString("\n")
		}
	}
	body.WriteString("\nEnter: select  N: new  D: delete  R: refresh  Q: quit\n")
	return newView(body.String())
}

func (m model) viewNewCampaign() tea.View {
	var body strings.Builder
	body.WriteString(titleStyle.Render("GopherQuest"))
	body.WriteString("\n\n")
	body.WriteString(menuStyle.Render("New Campaign"))
	body.WriteString("\n\n")
	for i, label := range formLabels {
		prefix := "  "
		if i == m.formField {
			prefix = "> "
		}
		value := m.formValues[i]
		if value == "" && i == titleField {
			value = placeholderStyle.Render("(e.g., Go Protocol Run)")
		}
		if i == m.formField {
			value += "_"
		}
		line := fmt.Sprintf("%s%-12s%s", prefix, label+":", value)
		if i == m.formField {
			body.WriteString(selectedStyle.Render(line))
		} else {
			body.WriteString(line)
		}
		body.WriteString("\n")
	}
	if m.formField == moduleField {
		body.WriteString("\n")
		body.WriteString(hintStyle.Render("Module presets (Left/Right):"))
		body.WriteString("\n")
		for i, preset := range modulePresets {
			line := "  " + preset
			if preset == m.formValues[moduleField] {
				line = "> " + preset
				body.WriteString(selectedStyle.Render(line))
			} else {
				body.WriteString(line)
			}
			if i < len(modulePresets)-1 {
				body.WriteString("\n")
			}
		}
		body.WriteString("\n")
	}
	if m.formField == difficultyField {
		body.WriteString("\n")
		body.WriteString(hintStyle.Render("Difficulty options (Left/Right):"))
		body.WriteString("\n")
		for i, difficulty := range difficultyPresets {
			line := "  " + difficulty
			if difficulty == m.formValues[difficultyField] {
				line = "> " + difficulty
				body.WriteString(selectedStyle.Render(line))
			} else {
				body.WriteString(line)
			}
			if i < len(difficultyPresets)-1 {
				body.WriteString("\n")
			}
		}
		body.WriteString("\n")
	}
	if m.loading {
		body.WriteString("\nCreating campaign...\n")
	}
	if m.err != nil {
		body.WriteString("\n")
		body.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		body.WriteString("\n")
	}
	body.WriteString("\nEnter: next/save  Up/Down: field  Left/Right: choose  Esc: campaigns  Ctrl+C: quit\n")
	return newView(body.String())
}

func (m model) viewGame() tea.View {
	var body strings.Builder
	body.WriteString(titleStyle.Render("GopherQuest"))
	body.WriteString("\n\n")
	if m.loading && m.campaign == nil {
		body.WriteString("Loading campaign...\n")
		return newView(body.String())
	}
	if m.err != nil {
		body.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		body.WriteString("\n\nPress q to quit.\n")
		return newView(body.String())
	}

	body.WriteString(sceneForStage(m.campaign.Stage, m.legFrame))
	body.WriteString("\n\n")
	body.WriteString(fmt.Sprintf("Campaign: %d\n", m.campaign.ID))
	body.WriteString(stageStyle.Render(fmt.Sprintf("Stage: %s", m.campaign.Stage)))
	body.WriteString(fmt.Sprintf("\nScore: %d\n", m.campaign.Score))
	if !m.campaign.ModuleInfo.IsNull() {
		info, _ := m.campaign.ModuleInfo.Get()
		body.WriteString(fmt.Sprintf("Latest module: %s\n", info.LatestVersion))
	}
	if !m.campaign.CurrentChallenge.IsNull() {
		challenge, _ := m.campaign.CurrentChallenge.Get()
		body.WriteString("\n")
		body.WriteString(promptStyle.Render(challenge.Prompt))
		body.WriteString("\n")
		body.WriteString(hintStyle.Render(fmt.Sprintf("Hint: %s", challenge.Hint)))
		body.WriteString("\n")
	}
	if m.status != "" {
		body.WriteString("\n")
		if m.statusGood {
			body.WriteString(successStyle.Render(m.status))
		} else {
			body.WriteString(errorStyle.Render(m.status))
		}
		body.WriteString("\n")
	}
	if m.campaign.Stage != api.StageComplete {
		if m.loading {
			body.WriteString("\nSubmitting...\n")
		} else {
			body.WriteString(fmt.Sprintf("\nAnswer: %s_\n", m.answer))
		}
	}
	if m.campaign.Stage == api.StageComplete {
		if m.returnToList {
			body.WriteString("\nEsc: campaigns  Q: quit\n")
		} else {
			body.WriteString("\nQ: quit\n")
		}
	} else if m.returnToList {
		body.WriteString("\nEsc: campaigns  Enter: submit  Backspace: delete  Q: quit\n")
	} else {
		body.WriteString("\nEnter: submit  Backspace: delete  Q: quit\n")
	}
	return newView(body.String())
}

func (m model) getCampaignCmd() tea.Cmd {
	return func() tea.Msg {
		response, err := m.client.GetCampaign(context.Background(), api.GetCampaignParams{ID: m.campaignID})
		if err != nil {
			return errorMsg{err: fmt.Errorf("get campaign: %w", err)}
		}
		campaign, ok := response.(*api.Campaign)
		if !ok {
			return errorMsg{err: fmt.Errorf("get campaign returned unexpected response %T", response)}
		}
		return campaignLoadedMsg{campaign: campaign}
	}
}

func (m model) listCampaignsCmd() tea.Cmd {
	return func() tea.Msg {
		response, err := m.client.ListCampaigns(context.Background())
		if err != nil {
			return errorMsg{err: fmt.Errorf("list campaigns: %w", err)}
		}
		campaigns, ok := response.(*api.ListCampaignsOKApplicationJSON)
		if !ok {
			return errorMsg{err: fmt.Errorf("list campaigns returned unexpected response %T", response)}
		}
		return campaignsLoadedMsg{campaigns: []api.Campaign(*campaigns)}
	}
}

func (m model) createCampaignCmd() tea.Cmd {
	request := &api.CreateCampaignRequest{
		PlayerName: m.formValues[playerField],
		Title:      m.formValues[titleField],
		ModulePath: api.ModulePath(m.formValues[moduleField]),
		Difficulty: api.Difficulty(m.formValues[difficultyField]),
	}
	return func() tea.Msg {
		response, err := m.client.CreateCampaign(context.Background(), request)
		if err != nil {
			return errorMsg{err: fmt.Errorf("create campaign: %w", err)}
		}
		campaign, ok := response.(*api.Campaign)
		if !ok {
			return errorMsg{err: fmt.Errorf("create campaign returned unexpected response %T", response)}
		}
		return campaignCreatedMsg{campaign: campaign}
	}
}

func (m model) deleteCampaignCmd(id int64) tea.Cmd {
	return func() tea.Msg {
		response, err := m.client.DeleteCampaign(context.Background(), api.DeleteCampaignParams{ID: id})
		if err != nil {
			return errorMsg{err: fmt.Errorf("delete campaign: %w", err)}
		}
		if _, ok := response.(*api.DeleteCampaignNoContent); !ok {
			return errorMsg{err: fmt.Errorf("delete campaign returned unexpected response %T", response)}
		}
		return campaignDeletedMsg{id: id}
	}
}

func (m model) submitAttemptCmd(answer string) tea.Cmd {
	return func() tea.Msg {
		response, err := m.client.SubmitAttempt(context.Background(), &api.AttemptRequest{Answer: answer}, api.SubmitAttemptParams{ID: m.campaignID})
		if err != nil {
			return errorMsg{err: fmt.Errorf("submit attempt: %w", err)}
		}
		result, ok := response.(*api.AttemptResult)
		if !ok {
			return errorMsg{err: fmt.Errorf("submit attempt returned unexpected response %T", response)}
		}
		return attemptLoadedMsg{result: result}
	}
}

func removeCampaign(campaigns []api.Campaign, id int64) []api.Campaign {
	for i, item := range campaigns {
		if item.ID != id {
			continue
		}
		return append(campaigns[:i], campaigns[i+1:]...)
	}
	return campaigns
}

func clampSelection(selected, length int) int {
	if length == 0 {
		return 0
	}
	if selected < 0 {
		return 0
	}
	if selected >= length {
		return length - 1
	}
	return selected
}

func newView(content string) tea.View {
	view := tea.NewView(content)
	view.AltScreen = true
	view.WindowTitle = "GopherQuest"
	return view
}

func animationTickCmd() tea.Cmd {
	return tea.Tick(350*time.Millisecond, func(time.Time) tea.Msg {
		return animationTickMsg{}
	})
}

func sceneForStage(stage api.Stage, legFrame bool) string {
	sprite := renderGopher(gopherStanding)
	if legFrame {
		sprite = renderGopher(gopherRunning)
	}
	return sprite + "\n" + trackForStage(stage)
}

func renderGopher(sprite string) string {
	lines := strings.Split(sprite, "\n")
	for i, line := range lines {
		lines[i] = gopherStyle.Render(line)
	}
	return strings.Join(lines, "\n")
}

func trackForStage(stage api.Stage) string {
	ground := func(value string) string { return groundStyle.Render(value) }
	pit := pitStyle.Render("______")
	finish := finishStyle.Render("===>FINISH")

	switch stage {
	case api.StageModule:
		return ground("================") + "   " + pit + "   " + ground("==============") + "   " + pit + "   " + ground("========") + "   " + pit + "   " + finish
	case api.StageErrors:
		return ground("==============") + "   " + pit + "   " + ground("========") + "   " + pit + "   " + finish
	case api.StageChannel:
		return ground("========") + "   " + pit + "   " + finish
	case api.StageComplete:
		return finish
	default:
		return ground("========") + "   " + pit + "   " + finish
	}
}

const (
	gopherStanding = "/\\_\n( o )\n /V\\\n/ \\"
	gopherRunning  = "/\\_\n( o )\n _V/\n / \\"
)

func removeLastRune(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return ""
	}
	return string(runes[:len(runes)-1])
}

func cycleModulePreset(value string, direction int) string {
	return cyclePreset(value, modulePresets, direction)
}

func cyclePreset(value string, presets []string, direction int) string {
	index := -1
	for i, preset := range presets {
		if preset == value {
			index = i
			break
		}
	}
	if index == -1 {
		if direction < 0 {
			return presets[len(presets)-1]
		}
		return presets[0]
	}

	index += direction
	if index < 0 {
		index = len(presets) - 1
	}
	if index >= len(presets) {
		index = 0
	}
	return presets[index]
}

func isModulePreset(value string) bool {
	for _, preset := range modulePresets {
		if preset == value {
			return true
		}
	}
	return false
}
