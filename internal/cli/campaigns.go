package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"gitlab.galvanize.com/cohorts/ent/sf/sdi-golang/sdi-go-2026-08-10/student-work/jiayang-wang/capstone/internal/api"
)

// Client is the generated API surface used by the CLI commands.
type Client interface {
	CreateCampaign(context.Context, *api.CreateCampaignRequest) (api.CreateCampaignRes, error)
	GetCampaign(context.Context, api.GetCampaignParams) (api.GetCampaignRes, error)
	SubmitAttempt(context.Context, *api.AttemptRequest, api.SubmitAttemptParams) (api.SubmitAttemptRes, error)
}

// NewRootCommand builds the Cobra command tree over a generated API client.
func NewRootCommand(client Client) *cobra.Command {
	root := &cobra.Command{
		Use:          "gopherquestctl",
		Short:        "Play GopherQuest from the terminal",
		SilenceUsage: true,
	}

	campaigns := &cobra.Command{
		Use:   "campaigns",
		Short: "Manage learning campaigns",
	}
	campaigns.AddCommand(newAddCommand(client), newGetCommand(client), newAnswerCommand(client))
	root.AddCommand(campaigns)
	return root
}

func newAddCommand(client Client) *cobra.Command {
	var player, title, modulePath, difficulty string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Create a campaign",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := client.CreateCampaign(cmd.Context(), &api.CreateCampaignRequest{
				PlayerName: player,
				Title:      title,
				ModulePath: api.ModulePath(modulePath),
				Difficulty: api.Difficulty(difficulty),
			})
			if err != nil {
				return fmt.Errorf("create campaign: %w", err)
			}
			campaign, ok := response.(*api.Campaign)
			if !ok {
				return fmt.Errorf("create campaign returned unexpected response %T", response)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Created campaign %d at stage %s\n", campaign.ID, campaign.Stage)
			return err
		},
	}
	cmd.Flags().StringVar(&player, "player", "", "player name")
	cmd.Flags().StringVar(&title, "title", "", "campaign title")
	cmd.Flags().StringVar(&modulePath, "module", "", "lower-case GitHub module path")
	cmd.Flags().StringVar(&difficulty, "difficulty", "beginner", "beginner, intermediate, or advanced")
	_ = cmd.MarkFlagRequired("player")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("module")
	return cmd
}

func newGetCommand(client Client) *cobra.Command {
	var id int64

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Show a campaign and its current challenge",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := client.GetCampaign(cmd.Context(), api.GetCampaignParams{ID: id})
			if err != nil {
				return fmt.Errorf("get campaign: %w", err)
			}
			campaign, ok := response.(*api.Campaign)
			if !ok {
				return fmt.Errorf("get campaign returned unexpected response %T", response)
			}
			return writeCampaign(cmd.OutOrStdout(), *campaign)
		},
	}
	cmd.Flags().Int64Var(&id, "id", 0, "campaign ID")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func newAnswerCommand(client Client) *cobra.Command {
	var id int64
	var answer string

	cmd := &cobra.Command{
		Use:   "answer",
		Short: "Submit an answer for the current challenge",
		RunE: func(cmd *cobra.Command, _ []string) error {
			response, err := client.SubmitAttempt(cmd.Context(), &api.AttemptRequest{Answer: answer}, api.SubmitAttemptParams{ID: id})
			if err != nil {
				return fmt.Errorf("submit answer: %w", err)
			}
			result, ok := response.(*api.AttemptResult)
			if !ok {
				return fmt.Errorf("submit answer returned unexpected response %T", response)
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\n", result.Message); err != nil {
				return err
			}
			return writeCampaign(cmd.OutOrStdout(), result.Campaign)
		},
	}
	cmd.Flags().Int64Var(&id, "id", 0, "campaign ID")
	cmd.Flags().StringVar(&answer, "answer", "", "challenge answer")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("answer")
	return cmd
}

func writeCampaign(w io.Writer, item api.Campaign) error {
	if _, err := fmt.Fprintf(w, "Campaign %d: %s\nPlayer: %s\nStage: %s\nScore: %d\n", item.ID, item.Title, item.PlayerName, item.Stage, item.Score); err != nil {
		return err
	}
	if !item.ModuleInfo.IsNull() {
		info, _ := item.ModuleInfo.Get()
		if _, err := fmt.Fprintf(w, "Latest module: %s\n", info.LatestVersion); err != nil {
			return err
		}
	}
	if !item.CurrentChallenge.IsNull() {
		challenge, _ := item.CurrentChallenge.Get()
		if _, err := fmt.Fprintf(w, "Challenge: %s\nHint: %s\n", challenge.Prompt, challenge.Hint); err != nil {
			return err
		}
	}
	return nil
}
