package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

var ErrAccountInvalid = errors.New("dados de conta invalidos")

type CreateAccountInput struct {
	ID       string
	UserID   string
	Currency string
	Scale    int32
}

type CreateAccountUseCase struct {
	accounts repository.AccountRepository
	clock    Clock
}

func NewCreateAccountUseCase(accounts repository.AccountRepository) *CreateAccountUseCase {
	return &CreateAccountUseCase{
		accounts: accounts,
		clock:    NewRealClock(),
	}
}

func NewCreateAccountUseCaseWithClock(accounts repository.AccountRepository, clock Clock) *CreateAccountUseCase {
	return &CreateAccountUseCase{
		accounts: accounts,
		clock:    clock,
	}
}

func (uc *CreateAccountUseCase) Execute(ctx context.Context, input CreateAccountInput) (*entity.Account, error) {
	if input.UserID == "" || input.Currency == "" {
		return nil, ErrAccountInvalid
	}

	if input.ID != "" {
		existing, err := uc.accounts.GetByID(ctx, input.ID)
		if err == nil && existing != nil {
			return existing, nil
		}
		if err != nil && err != repository.ErrNotFound {
			return nil, err
		}
	}

	id := input.ID
	if id == "" {
		id = uuid.NewString()
	}

	now := uc.clock.Now().UTC()
	account := &entity.Account{
		ID:        id,
		UserID:    input.UserID,
		Currency:  input.Currency,
		Scale:     input.Scale,
		Status:    entity.AccountStatusActive,
		CreatedAt: now,
	}

	if err := uc.accounts.Create(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}
