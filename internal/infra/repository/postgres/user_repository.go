package postgres

import (
	"context"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

type UserRepository struct {
	db dbExecutor
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: pool}
}

func NewUserRepositoryWithTx(tx dbExecutor) *UserRepository {
	return &UserRepository{db: tx}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, external_id, email, full_name, status, user_type, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		user.ID,
		nullIfEmpty(user.ExternalID),
		user.Email,
		user.FullName,
		user.Status,
		user.UserType,
		nullIfEmpty(user.PasswordHash),
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, external_id, email, full_name, status, user_type, password_hash, created_at, updated_at
		  FROM users
		 WHERE id = $1`, id)
	var user entity.User
	var externalID *string
	var passwordHash *string
	if err := row.Scan(
		&user.ID,
		&externalID,
		&user.Email,
		&user.FullName,
		&user.Status,
		&user.UserType,
		&passwordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	if externalID != nil {
		user.ExternalID = *externalID
	}
	if passwordHash != nil {
		user.PasswordHash = *passwordHash
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, external_id, email, full_name, status, user_type, password_hash, created_at, updated_at
		  FROM users
		 WHERE email = $1`, email)
	var user entity.User
	var externalID *string
	var passwordHash *string
	if err := row.Scan(
		&user.ID,
		&externalID,
		&user.Email,
		&user.FullName,
		&user.Status,
		&user.UserType,
		&passwordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if externalID != nil {
		user.ExternalID = *externalID
	}
	if passwordHash != nil {
		user.PasswordHash = *passwordHash
	}
	return &user, nil
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id string, status entity.UserStatus) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users
		   SET status = $1, updated_at = NOW()
		 WHERE id = $2`, status, id)
	return err
}

func (r *UserRepository) Search(ctx context.Context, filter repository.UserSearchFilter) ([]*entity.User, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	var builder strings.Builder
	builder.WriteString(`
		SELECT id, external_id, email, full_name, status, user_type, password_hash, created_at, updated_at
		  FROM users
		 WHERE 1=1`)
	args := []any{}
	argIndex := 1

	if filter.ID != "" {
		builder.WriteString(" AND id = $" + strconv.Itoa(argIndex))
		args = append(args, filter.ID)
		argIndex++
	}
	if filter.Email != "" {
		builder.WriteString(" AND LOWER(email) = LOWER($" + strconv.Itoa(argIndex) + ")")
		args = append(args, filter.Email)
		argIndex++
	}
	if filter.ExternalID != "" {
		builder.WriteString(" AND external_id = $" + strconv.Itoa(argIndex))
		args = append(args, filter.ExternalID)
		argIndex++
	}

	builder.WriteString(" ORDER BY created_at DESC LIMIT $" + strconv.Itoa(argIndex))
	args = append(args, limit)

	rows, err := r.db.Query(ctx, builder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		var user entity.User
		var externalID *string
		var passwordHash *string
		if err := rows.Scan(
			&user.ID,
			&externalID,
			&user.Email,
			&user.FullName,
			&user.Status,
			&user.UserType,
			&passwordHash,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if externalID != nil {
			user.ExternalID = *externalID
		}
		if passwordHash != nil {
			user.PasswordHash = *passwordHash
		}
		users = append(users, &user)
	}
	return users, rows.Err()
}

var _ repository.UserRepository = (*UserRepository)(nil)
