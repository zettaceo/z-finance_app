package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type RoleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{db: pool}
}

func (r *RoleRepository) ListAll(ctx context.Context) ([]*entity.Role, error) {
	rows, err := r.db.Query(ctx, `
		SELECT code, description, created_at
		  FROM roles
		 ORDER BY code ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.Role
	for rows.Next() {
		var item entity.Role
		if err := rows.Scan(&item.Code, &item.Description, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, rows.Err()
}

func (r *RoleRepository) Create(ctx context.Context, role *entity.Role) error {
	if role == nil {
		return repository.ErrInvalidState
	}
	if role.CreatedAt.IsZero() {
		role.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO roles (code, description, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (code) DO NOTHING`,
		role.Code,
		nullIfEmpty(role.Description),
		role.CreatedAt,
	)
	return err
}

var _ repository.RoleRepository = (*RoleRepository)(nil)
