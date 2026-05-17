package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type RegulatoryProfileRepository struct {
	db *pgxpool.Pool
}

func NewRegulatoryProfileRepository(pool *pgxpool.Pool) *RegulatoryProfileRepository {
	return &RegulatoryProfileRepository{db: pool}
}

func (r *RegulatoryProfileRepository) GetByUserID(ctx context.Context, userID string) (*entity.RegulatoryProfile, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, jurisdiction_code, jurisdiction_risk, aml_tier, travel_rule_required,
		       sanctions_screening_required, created_at, updated_at
		  FROM regulatory_profiles
		 WHERE user_id = $1`, userID)
	var item entity.RegulatoryProfile
	if err := row.Scan(
		&item.ID,
		&item.UserID,
		&item.JurisdictionCode,
		&item.JurisdictionRisk,
		&item.AMLTier,
		&item.TravelRuleRequired,
		&item.SanctionsScreeningRequired,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *RegulatoryProfileRepository) Upsert(ctx context.Context, profile *entity.RegulatoryProfile) error {
	if profile == nil {
		return repository.ErrInvalidState
	}
	now := time.Now().UTC()
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = now
	}
	profile.UpdatedAt = now
	if profile.ID == "" {
		profile.ID = profile.UserID
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO regulatory_profiles (
			id, user_id, jurisdiction_code, jurisdiction_risk, aml_tier,
			travel_rule_required, sanctions_screening_required, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		ON CONFLICT (user_id) DO UPDATE SET
			jurisdiction_code = EXCLUDED.jurisdiction_code,
			jurisdiction_risk = EXCLUDED.jurisdiction_risk,
			aml_tier = EXCLUDED.aml_tier,
			travel_rule_required = EXCLUDED.travel_rule_required,
			sanctions_screening_required = EXCLUDED.sanctions_screening_required,
			updated_at = EXCLUDED.updated_at`,
		profile.ID,
		profile.UserID,
		profile.JurisdictionCode,
		profile.JurisdictionRisk,
		profile.AMLTier,
		profile.TravelRuleRequired,
		profile.SanctionsScreeningRequired,
		profile.CreatedAt,
		profile.UpdatedAt,
	)
	return err
}

var _ repository.RegulatoryProfileRepository = (*RegulatoryProfileRepository)(nil)
