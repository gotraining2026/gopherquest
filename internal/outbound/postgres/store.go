package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.galvanize.com/cohorts/ent/sf/sdi-golang/sdi-go-2026-08-10/student-work/jiayang-wang/capstone/internal/campaign"
)

// Store adapts generated PostgreSQL queries to campaign domain values.
type Store struct {
	queries *Queries
}

// NewStore builds a campaign store over a PostgreSQL connection pool.
func NewStore(db *sql.DB) *Store {
	return &Store{queries: New(db)}
}

// Create persists a new campaign.
func (s *Store) Create(ctx context.Context, input campaign.NewCampaign) (campaign.Campaign, error) {
	row, err := s.queries.CreateCampaign(ctx, CreateCampaignParams{
		PlayerName: input.PlayerName,
		Title:      input.Title,
		ModulePath: input.ModulePath,
		Difficulty: string(input.Difficulty),
	})
	if err != nil {
		return campaign.Campaign{}, fmt.Errorf("create campaign row: %w", err)
	}
	return toCampaign(row), nil
}

// List returns all persisted campaigns without live enrichment.
func (s *Store) List(ctx context.Context) ([]campaign.Campaign, error) {
	rows, err := s.queries.ListCampaigns(ctx)
	if err != nil {
		return nil, fmt.Errorf("list campaign rows: %w", err)
	}

	items := make([]campaign.Campaign, 0, len(rows))
	for _, row := range rows {
		items = append(items, toCampaign(row))
	}
	return items, nil
}

// Get returns one persisted campaign without live enrichment.
func (s *Store) Get(ctx context.Context, id int64) (campaign.Campaign, error) {
	row, err := s.queries.GetCampaign(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return campaign.Campaign{}, campaign.ErrNotFound
	}
	if err != nil {
		return campaign.Campaign{}, fmt.Errorf("get campaign row %d: %w", id, err)
	}
	return toCampaign(row), nil
}

// Update replaces editable campaign fields before the first stage completes.
func (s *Store) Update(ctx context.Context, id int64, input campaign.UpdateCampaign) (campaign.Campaign, error) {
	row, err := s.queries.UpdateCampaign(ctx, UpdateCampaignParams{
		ID:         id,
		PlayerName: input.PlayerName,
		Title:      input.Title,
		ModulePath: input.ModulePath,
		Difficulty: string(input.Difficulty),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return campaign.Campaign{}, s.classifyExistingCampaign(ctx, id)
	}
	if err != nil {
		return campaign.Campaign{}, fmt.Errorf("update campaign row %d: %w", id, err)
	}
	return toCampaign(row), nil
}

// Delete removes a campaign and reports not-found when no row was removed.
func (s *Store) Delete(ctx context.Context, id int64) error {
	rows, err := s.queries.DeleteCampaign(ctx, id)
	if err != nil {
		return fmt.Errorf("delete campaign row %d: %w", id, err)
	}
	if rows == 0 {
		return campaign.ErrNotFound
	}
	return nil
}

// Advance changes stage and score in one expected-stage update.
func (s *Store) Advance(ctx context.Context, id int64, expected, next campaign.Stage, points int) (campaign.Campaign, error) {
	row, err := s.queries.AdvanceCampaign(ctx, AdvanceCampaignParams{
		ID:      id,
		Stage:   string(next),
		Score:   int32(points),
		Stage_2: string(expected),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return campaign.Campaign{}, s.classifyExistingCampaign(ctx, id)
	}
	if err != nil {
		return campaign.Campaign{}, fmt.Errorf("advance campaign row %d: %w", id, err)
	}
	return toCampaign(row), nil
}

func (s *Store) classifyExistingCampaign(ctx context.Context, id int64) error {
	_, err := s.Get(ctx, id)
	if errors.Is(err, campaign.ErrNotFound) {
		return campaign.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("classify campaign %d: %w", id, err)
	}
	return campaign.ErrWrongStage
}

func toCampaign(row Campaign) campaign.Campaign {
	return campaign.Campaign{
		ID:         row.ID,
		PlayerName: row.PlayerName,
		Title:      row.Title,
		ModulePath: row.ModulePath,
		Difficulty: campaign.Difficulty(row.Difficulty),
		Stage:      campaign.Stage(row.Stage),
		Score:      int(row.Score),
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}
