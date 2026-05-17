package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type RoleSeparationRepository struct {
	db *pgxpool.Pool
}

func NewRoleSeparationRepository(pool *pgxpool.Pool) *RoleSeparationRepository {
	return &RoleSeparationRepository{db: pool}
}

func (r *RoleSeparationRepository) ListAll(ctx context.Context) ([]*entity.RoleSeparationRule, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, role_code_a, role_code_b, reason, created_at
		  FROM role_separation_rules
		 ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.RoleSeparationRule
	for rows.Next() {
		var item entity.RoleSeparationRule
		if err := rows.Scan(&item.ID, &item.RoleCodeA, &item.RoleCodeB, &item.Reason, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, rows.Err()
}

func (r *RoleSeparationRepository) Create(ctx context.Context, rule *entity.RoleSeparationRule) error {
	if rule == nil {
		return repository.ErrInvalidState
	}
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO role_separation_rules (id, role_code_a, role_code_b, reason, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (role_code_a, role_code_b) DO NOTHING`,
		rule.ID,
		rule.RoleCodeA,
		rule.RoleCodeB,
		nullIfEmpty(rule.Reason),
		rule.CreatedAt,
	)
	return err
}

func (r *RoleSeparationRepository) Remove(ctx context.Context, roleCodeA, roleCodeB string) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM role_separation_rules
		 WHERE role_code_a = $1
		   AND role_code_b = $2`, roleCodeA, roleCodeB)
	return err
}

func (r *RoleSeparationRepository) HasConflict(ctx context.Context, userID, roleCode string) (bool, error) {
	row := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			  FROM role_separation_rules rs
			  JOIN user_roles ur
			    ON ur.user_id = $1
			   AND (
			        (rs.role_code_a = $2 AND ur.role_code = rs.role_code_b)
			     OR (rs.role_code_b = $2 AND ur.role_code = rs.role_code_a)
			   )
		)`, userID, roleCode)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

var _ repository.RoleSeparationRepository = (*RoleSeparationRepository)(nil)
