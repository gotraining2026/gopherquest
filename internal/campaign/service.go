package campaign

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const queryTimeout = 3 * time.Second

// Service contains campaign validation, persistence orchestration, and live
// enrichment without depending on transport or generated API types.
type Service struct {
	store        Store
	moduleLookup ModuleLookup
}

// NewService builds a campaign service over its consumed interfaces.
func NewService(store Store, moduleLookup ModuleLookup) *Service {
	return &Service{store: store, moduleLookup: moduleLookup}
}

// Create validates and persists a campaign with normalized editable fields.
func (s *Service) Create(ctx context.Context, input NewCampaign) (Campaign, error) {
	input = normalizeNew(input)
	if err := input.Validate(); err != nil {
		return Campaign{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	created, err := s.store.Create(ctx, input)
	if err != nil {
		return Campaign{}, fmt.Errorf("create campaign: %w", err)
	}
	return withChallenge(created), nil
}

// List returns persisted campaigns without calling the external provider.
func (s *Service) List(ctx context.Context) ([]Campaign, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	items, err := s.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list campaigns: %w", err)
	}
	for i := range items {
		items[i] = withChallenge(items[i])
	}
	return items, nil
}

// Get returns one campaign with live module enrichment.
func (s *Service) Get(ctx context.Context, id int64) (Campaign, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	item, err := s.store.Get(ctx, id)
	if err != nil {
		return Campaign{}, fmt.Errorf("get campaign %d: %w", id, err)
	}
	if s.moduleLookup == nil {
		return Campaign{}, fmt.Errorf("get campaign %d: %w", id, ErrProvider)
	}

	moduleInfo, err := s.moduleLookup.Latest(ctx, item.ModulePath)
	if err != nil {
		return Campaign{}, fmt.Errorf("enrich campaign %d: %w", id, err)
	}
	item.ModuleInfo = &moduleInfo
	return withChallenge(item), nil
}

// Update validates and replaces editable fields before the first stage.
func (s *Service) Update(ctx context.Context, id int64, input UpdateCampaign) (Campaign, error) {
	input = normalizeUpdate(input)
	if err := input.Validate(); err != nil {
		return Campaign{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	updated, err := s.store.Update(ctx, id, input)
	if err != nil {
		return Campaign{}, fmt.Errorf("update campaign %d: %w", id, err)
	}
	return withChallenge(updated), nil
}

// Delete removes a campaign.
func (s *Service) Delete(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	if err := s.store.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete campaign %d: %w", id, err)
	}
	return nil
}

func normalizeNew(input NewCampaign) NewCampaign {
	input.PlayerName = strings.TrimSpace(input.PlayerName)
	input.Title = strings.TrimSpace(input.Title)
	input.ModulePath = strings.TrimSpace(input.ModulePath)
	return input
}

func normalizeUpdate(input UpdateCampaign) UpdateCampaign {
	input.PlayerName = strings.TrimSpace(input.PlayerName)
	input.Title = strings.TrimSpace(input.Title)
	input.ModulePath = strings.TrimSpace(input.ModulePath)
	return input
}

func withChallenge(item Campaign) Campaign {
	item.CurrentChallenge = ChallengeFor(item.Stage)
	return item
}
