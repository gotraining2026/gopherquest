package inbound

import (
	"context"
	"errors"

	"gitlab.galvanize.com/cohorts/ent/sf/sdi-golang/sdi-go-2026-08-10/student-work/jiayang-wang/capstone/internal/api"
	"gitlab.galvanize.com/cohorts/ent/sf/sdi-golang/sdi-go-2026-08-10/student-work/jiayang-wang/capstone/internal/campaign"
)

// Handler implements the generated API handler over the campaign service.
type Handler struct {
	api.UnimplementedHandler
	service *campaign.Service
}

// NewHandler builds an API handler over a campaign service.
func NewHandler(service *campaign.Service) *Handler {
	return &Handler{service: service}
}

// CreateCampaign handles POST /campaigns.
func (h *Handler) CreateCampaign(ctx context.Context, req *api.CreateCampaignRequest) (api.CreateCampaignRes, error) {
	created, err := h.service.Create(ctx, campaign.NewCampaign{
		PlayerName: req.PlayerName,
		Title:      req.Title,
		ModulePath: string(req.ModulePath),
		Difficulty: campaign.Difficulty(req.Difficulty),
	})
	if err != nil {
		return nil, err
	}
	response := toWire(created)
	return &response, nil
}

// DeleteCampaign handles DELETE /campaigns/{id}.
func (h *Handler) DeleteCampaign(ctx context.Context, params api.DeleteCampaignParams) (api.DeleteCampaignRes, error) {
	if err := h.service.Delete(ctx, params.ID); err != nil {
		return nil, err
	}
	return &api.DeleteCampaignNoContent{}, nil
}

// GetCampaign handles GET /campaigns/{id}.
func (h *Handler) GetCampaign(ctx context.Context, params api.GetCampaignParams) (api.GetCampaignRes, error) {
	item, err := h.service.Get(ctx, params.ID)
	if err != nil {
		return nil, err
	}
	response := toWire(item)
	return &response, nil
}

// ListCampaigns handles GET /campaigns.
func (h *Handler) ListCampaigns(ctx context.Context) (api.ListCampaignsRes, error) {
	items, err := h.service.List(ctx)
	if err != nil {
		return nil, err
	}

	response := make(api.ListCampaignsOKApplicationJSON, 0, len(items))
	for _, item := range items {
		response = append(response, toWire(item))
	}
	return &response, nil
}

// UpdateCampaign handles PUT /campaigns/{id}.
func (h *Handler) UpdateCampaign(ctx context.Context, req *api.UpdateCampaignRequest, params api.UpdateCampaignParams) (api.UpdateCampaignRes, error) {
	updated, err := h.service.Update(ctx, params.ID, campaign.UpdateCampaign{
		PlayerName: req.PlayerName,
		Title:      req.Title,
		ModulePath: string(req.ModulePath),
		Difficulty: campaign.Difficulty(req.Difficulty),
	})
	if err != nil {
		return nil, err
	}
	response := toWire(updated)
	return &response, nil
}

// NewError maps domain errors to stable API status responses.
func (h *Handler) NewError(_ context.Context, err error) *api.InternalErrorStatusCode {
	status := statusForError(err)
	message := "internal server error"
	switch status {
	case 400:
		message = "invalid campaign"
	case 404:
		message = "campaign not found"
	case 409:
		message = "campaign state conflict"
	case 502:
		message = "module provider unavailable"
	}

	return &api.InternalErrorStatusCode{
		StatusCode: status,
		Response: api.ErrorResponse{
			Error: message,
		},
	}
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, campaign.ErrInvalid):
		return 400
	case errors.Is(err, campaign.ErrNotFound):
		return 404
	case errors.Is(err, campaign.ErrWrongStage):
		return 409
	case errors.Is(err, campaign.ErrProvider):
		return 502
	default:
		return 500
	}
}

func toWire(item campaign.Campaign) api.Campaign {
	response := api.Campaign{
		ID:         item.ID,
		PlayerName: item.PlayerName,
		Title:      item.Title,
		ModulePath: item.ModulePath,
		Difficulty: api.Difficulty(item.Difficulty),
		Stage:      api.Stage(item.Stage),
		Score:      int32(item.Score),
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
	}

	if item.ModuleInfo == nil {
		response.ModuleInfo.SetToNull()
	} else {
		response.ModuleInfo.SetTo(api.ModuleInfo{
			LatestVersion: item.ModuleInfo.LatestVersion,
			PublishedAt:   item.ModuleInfo.PublishedAt,
		})
	}

	if item.CurrentChallenge == nil {
		response.CurrentChallenge.SetToNull()
	} else {
		response.CurrentChallenge.SetTo(api.Challenge{
			Stage:  api.Stage(item.CurrentChallenge.Stage),
			Prompt: item.CurrentChallenge.Prompt,
			Hint:   item.CurrentChallenge.Hint,
		})
	}

	return response
}
