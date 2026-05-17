package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

var ErrUserInvalid = errors.New("dados de usuario invalidos")

type CreateUserInput struct {
	ID       string
	Email    string
	FullName string
	Password string
	UserType entity.UserType
}

type CreateUserUseCase struct {
	users repository.UserRepository
	clock Clock
}

func NewCreateUserUseCase(users repository.UserRepository) *CreateUserUseCase {
	return &CreateUserUseCase{
		users: users,
		clock: NewRealClock(),
	}
}

func NewCreateUserUseCaseWithClock(users repository.UserRepository, clock Clock) *CreateUserUseCase {
	return &CreateUserUseCase{
		users: users,
		clock: clock,
	}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*entity.User, error) {
	if input.Email == "" || input.FullName == "" {
		return nil, ErrUserInvalid
	}

	existing, err := uc.users.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	id := input.ID
	if id == "" {
		id = uuid.NewString()
	}

	passwordHash := ""
	if input.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		passwordHash = string(hash)
	}

	now := uc.clock.Now().UTC()
	user := &entity.User{
		ID:           id,
		Email:        input.Email,
		FullName:     input.FullName,
		Status:       entity.UserStatusActive,
		UserType:     defaultUserType(input.UserType),
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func defaultUserType(value entity.UserType) entity.UserType {
	if value == "" {
		return entity.UserTypePF
	}
	return value
}
