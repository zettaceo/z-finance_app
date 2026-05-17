package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
)

type UserSettingsRepository struct {
	pool *pgxpool.Pool
}

func NewUserSettingsRepository(pool *pgxpool.Pool) *UserSettingsRepository {
	return &UserSettingsRepository{pool: pool}
}

func (r *UserSettingsRepository) Upsert(ctx context.Context, settings *entity.UserSettings) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_settings (
			user_id, conversion_priority, allow_crypto_to_fiat, auto_convert_pix_in,
			pix_in_target_asset, pix_in_percentage, auto_convert_enabled, auto_convert_asset,
			auto_convert_min_amount, fallback_asset, ux_mode, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE
		SET conversion_priority = EXCLUDED.conversion_priority,
		    allow_crypto_to_fiat = EXCLUDED.allow_crypto_to_fiat,
		    auto_convert_pix_in = EXCLUDED.auto_convert_pix_in,
		    pix_in_target_asset = EXCLUDED.pix_in_target_asset,
		    pix_in_percentage = EXCLUDED.pix_in_percentage,
		    auto_convert_enabled = EXCLUDED.auto_convert_enabled,
		    auto_convert_asset = EXCLUDED.auto_convert_asset,
		    auto_convert_min_amount = EXCLUDED.auto_convert_min_amount,
		    fallback_asset = EXCLUDED.fallback_asset,
		    ux_mode = EXCLUDED.ux_mode,
		    updated_at = NOW()`,
		settings.UserID,
		settings.ConversionPriority,
		settings.AllowCryptoToFiat,
		settings.AutoConvertPixIn,
		nullIfEmpty(settings.PixInTargetAsset),
		settings.PixInPercentage,
		settings.AutoConvertEnabled,
		nullIfEmpty(settings.AutoConvertAsset),
		settings.AutoConvertMinAmount,
		nullIfEmpty(settings.FallbackAsset),
		nullIfEmpty(settings.UXMode),
	)
	return err
}

func (r *UserSettingsRepository) GetByUserID(ctx context.Context, userID string) (*entity.UserSettings, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT user_id, conversion_priority, allow_crypto_to_fiat, auto_convert_pix_in,
		       pix_in_target_asset, pix_in_percentage, auto_convert_enabled, auto_convert_asset,
		       auto_convert_min_amount, fallback_asset, ux_mode, created_at, updated_at
		  FROM user_settings
		 WHERE user_id = $1`, userID)

	var settings entity.UserSettings
	var targetAsset *string
	var autoConvertAsset *string
	var fallbackAsset *string
	if err := row.Scan(
		&settings.UserID,
		&settings.ConversionPriority,
		&settings.AllowCryptoToFiat,
		&settings.AutoConvertPixIn,
		&targetAsset,
		&settings.PixInPercentage,
		&settings.AutoConvertEnabled,
		&autoConvertAsset,
		&settings.AutoConvertMinAmount,
		&fallbackAsset,
		&settings.UXMode,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if targetAsset != nil {
		settings.PixInTargetAsset = *targetAsset
	}
	if fallbackAsset != nil {
		settings.FallbackAsset = *fallbackAsset
	}
	if autoConvertAsset != nil {
		settings.AutoConvertAsset = *autoConvertAsset
	}

	return &settings, nil
}
