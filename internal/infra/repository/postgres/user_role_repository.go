package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type UserRoleRepository struct {
	db *pgxpool.Pool
}

func NewUserRoleRepository(pool *pgxpool.Pool) *UserRoleRepository {
	return &UserRoleRepository{db: pool}
}

func (r *UserRoleRepository) ListByUser(ctx context.Context, userID string) ([]*entity.UserRole, error) {
	rows, err := r.db.Query(ctx, `
		SELECT user_id, role_code, granted_by, granted_at
		  FROM user_roles
		 WHERE user_id = $1
		 ORDER BY role_code ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.UserRole
	for rows.Next() {
		var item entity.UserRole
		var grantedBy *string
		if err := rows.Scan(&item.UserID, &item.RoleCode, &grantedBy, &item.GrantedAt); err != nil {
			return nil, err
		}
		if grantedBy != nil {
			item.GrantedBy = *grantedBy
		}
		items = append(items, &item)
	}
	return items, rows.Err()
}

func (r *UserRoleRepository) Assign(ctx context.Context, userRole *entity.UserRole) error {
	if userRole == nil {
		return repository.ErrInvalidState
	}
	if userRole.GrantedAt.IsZero() {
		userRole.GrantedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_code, granted_by, granted_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, role_code) DO NOTHING`,
		userRole.UserID,
		userRole.RoleCode,
		nullIfEmpty(userRole.GrantedBy),
		userRole.GrantedAt,
	)
	return err
}

func (r *UserRoleRepository) Remove(ctx context.Context, userID, roleCode string) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM user_roles
		 WHERE user_id = $1
		   AND role_code = $2`, userID, roleCode)
	return err
}

var _ repository.UserRoleRepository = (*UserRoleRepository)(nil)
