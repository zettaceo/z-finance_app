package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/entity"
)

var (
	ErrPixKeyInvalidType   = errors.New("tipo de chave pix invalido")
	ErrPixKeyInvalidFormat = errors.New("formato de chave pix invalido")
	ErrPixKeyDuplicate     = errors.New("chave pix ja cadastrada")
)

type RegisterPixKeyInput struct {
	UserID    string
	AccountID string
	Type      string
	Key       string
}

type RegisterPixKeyUseCase struct {
	uow ports.UnitOfWork
	clock Clock
}

func NewRegisterPixKeyUseCase(uow ports.UnitOfWork) *RegisterPixKeyUseCase {
	return &RegisterPixKeyUseCase{uow: uow, clock: NewRealClock()}
}

func NewRegisterPixKeyUseCaseWithClock(uow ports.UnitOfWork, clock Clock) *RegisterPixKeyUseCase {
	return &RegisterPixKeyUseCase{uow: uow, clock: clock}
}

func (uc *RegisterPixKeyUseCase) Execute(ctx context.Context, input RegisterPixKeyInput) (*entity.PixKey, error) {
	if input.UserID == "" || input.AccountID == "" || input.Type == "" || input.Key == "" {
		return nil, ErrPixKeyInvalidFormat
	}
	keyType := entity.PixKeyType(strings.ToUpper(input.Type))
	if !isValidPixKeyType(keyType) {
		return nil, ErrPixKeyInvalidType
	}
	if !isValidPixKeyFormat(keyType, input.Key) {
		return nil, ErrPixKeyInvalidFormat
	}

	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = uowTx.Rollback(ctx)
	}()

	account, err := uowTx.AccountRepository().GetByIDForUpdate(ctx, input.AccountID)
	if err != nil {
		return nil, err
	}
	if account.UserID != input.UserID || account.Status != entity.AccountStatusActive {
		return nil, ErrAccountInactive
	}

	existing, err := uowTx.PixKeyRepository().GetByKey(ctx, input.Key)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrPixKeyDuplicate
	}

	now := uc.clock.Now().UTC()
	pixKey := &entity.PixKey{
		ID:        uuid.NewString(),
		UserID:    input.UserID,
		AccountID: input.AccountID,
		Type:      keyType,
		Key:       input.Key,
		CreatedAt: now,
	}
	if err := uowTx.PixKeyRepository().Create(ctx, pixKey); err != nil {
		return nil, err
	}

	if err := appendAudit(ctx, uowTx, uc.clock, input.UserID, "PIX_KEY_REGISTER", "pix_key", pixKey.ID, map[string]any{
		"type": pixKey.Type,
	}); err != nil {
		return nil, err
	}

	if err := uowTx.Commit(ctx); err != nil {
		return nil, err
	}
	return pixKey, nil
}

func isValidPixKeyType(t entity.PixKeyType) bool {
	switch t {
	case entity.PixKeyTypeCPF,
		entity.PixKeyTypeEmail,
		entity.PixKeyTypePhone,
		entity.PixKeyTypeEVP:
		return true
	default:
		return false
	}
}

func isValidPixKeyFormat(t entity.PixKeyType, value string) bool {
	trimmed := strings.TrimSpace(value)
	switch t {
	case entity.PixKeyTypeCPF:
		return regexp.MustCompile(`^\d{11}$`).MatchString(trimmed)
	case entity.PixKeyTypeEmail:
		return strings.Contains(trimmed, "@") && strings.Contains(trimmed, ".")
	case entity.PixKeyTypePhone:
		return regexp.MustCompile(`^\+?\d{10,15}$`).MatchString(trimmed)
	case entity.PixKeyTypeEVP:
		_, err := uuid.Parse(trimmed)
		return err == nil
	default:
		return false
	}
}

func appendAudit(ctx context.Context, uowTx ports.UnitOfWorkTx, clock Clock, userID, action, entityType, entityID string, data any) error {
	var payload json.RawMessage
	if data != nil {
		encoded, err := json.Marshal(data)
		if err != nil {
			return err
		}
		payload = encoded
	}
	logEntry := &entity.AuditLog{
		ID:         uuid.NewString(),
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Data:       payload,
		CreatedAt:  clock.Now().UTC(),
	}
	return uowTx.AuditLogRepository().Append(ctx, logEntry)
}

