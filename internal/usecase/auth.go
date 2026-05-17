package usecase

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"z-finance-api/internal/infra/observability"
	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

var ErrAuthInvalidCredentials = errors.New("credenciais invalidas")
var ErrAuthRateLimited = errors.New("muitas tentativas")
var ErrAuthTokenRevoked = errors.New("refresh token revogado")

type LoginInput struct {
	Email    string
	Password string
	IP       string
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	UserID       string
}

type LoginUseCase struct {
	uow          ports.UnitOfWork
	users        repository.UserRepository
	tokens       *TokenService
	refreshTTL   time.Duration
	maxAttempts  int64
	window       time.Duration
	clock        Clock
}

func NewLoginUseCase(uow ports.UnitOfWork, users repository.UserRepository, tokens *TokenService, refreshTTL time.Duration, maxAttempts int64, window time.Duration) *LoginUseCase {
	return &LoginUseCase{
		uow:         uow,
		users:       users,
		tokens:      tokens,
		refreshTTL:  refreshTTL,
		maxAttempts: maxAttempts,
		window:      window,
		clock:       NewRealClock(),
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginResult, error) {
	ctx, end := observability.StartUseCaseSpan(ctx, "Login")
	defer end()

	if input.Email == "" || input.Password == "" {
		return nil, ErrAuthInvalidCredentials
	}

	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = uowTx.Rollback(ctx) }()

	since := uc.clock.Now().UTC().Add(-uc.window)
	if uc.maxAttempts > 0 {
		failures, err := uowTx.LoginAuditRepository().CountRecentFailures(ctx, input.Email, input.IP, since)
		if err != nil {
			return nil, err
		}
		if failures >= uc.maxAttempts {
			_ = uowTx.LoginAuditRepository().Append(ctx, &entity.LoginAudit{
				ID:        uuid.NewString(),
				Email:     input.Email,
				IP:        input.IP,
				Success:   false,
				Reason:    "rate_limited",
				CreatedAt: uc.clock.Now().UTC(),
			})
			_ = uowTx.Commit(ctx)
			return nil, ErrAuthRateLimited
		}
	}

	user, err := uc.users.GetByEmail(ctx, input.Email)
	if err != nil || user == nil {
		_ = uc.appendLoginAudit(ctx, uowTx, "", input, false, "invalid_credentials")
		_ = uowTx.Commit(ctx)
		return nil, ErrAuthInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		_ = uc.appendLoginAudit(ctx, uowTx, user.ID, input, false, "invalid_credentials")
		_ = uowTx.Commit(ctx)
		return nil, ErrAuthInvalidCredentials
	}

	accessToken, expiresAt, err := uc.tokens.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshHash, err := uc.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	refreshEntity := &entity.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: uc.clock.Now().UTC().Add(uc.refreshTTL),
		CreatedAt: uc.clock.Now().UTC(),
	}
	if err := uowTx.RefreshTokenRepository().Create(ctx, refreshEntity); err != nil {
		return nil, err
	}

	_ = uc.appendLoginAudit(ctx, uowTx, user.ID, input, true, "success")

	if err := uowTx.Commit(ctx); err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		UserID:       user.ID,
	}, nil
}

type RefreshInput struct {
	RefreshToken string
	IP           string
}

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	UserID       string
}

type RefreshTokenUseCase struct {
	uow        ports.UnitOfWork
	tokens     *TokenService
	refreshTTL time.Duration
	clock      Clock
}

func NewRefreshTokenUseCase(uow ports.UnitOfWork, tokens *TokenService, refreshTTL time.Duration) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		uow:        uow,
		tokens:     tokens,
		refreshTTL: refreshTTL,
		clock:      NewRealClock(),
	}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, input RefreshInput) (*RefreshResult, error) {
	ctx, end := observability.StartUseCaseSpan(ctx, "RefreshToken")
	defer end()

	token := strings.TrimSpace(input.RefreshToken)
	if token == "" {
		return nil, ErrAuthInvalidCredentials
	}

	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = uowTx.Rollback(ctx) }()

	hash := hashFromToken(token)

	stored, err := uowTx.RefreshTokenRepository().GetByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if stored.RevokedAt != nil || uc.clock.Now().UTC().After(stored.ExpiresAt) {
		return nil, ErrAuthTokenRevoked
	}

	accessToken, expiresAt, err := uc.tokens.GenerateAccessToken(stored.UserID)
	if err != nil {
		return nil, err
	}

	newRefreshToken, newHash, err := uc.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	newEntity := &entity.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    stored.UserID,
		TokenHash: newHash,
		ExpiresAt: uc.clock.Now().UTC().Add(uc.refreshTTL),
		CreatedAt: uc.clock.Now().UTC(),
	}
	if err := uowTx.RefreshTokenRepository().Create(ctx, newEntity); err != nil {
		return nil, err
	}
	if err := uowTx.RefreshTokenRepository().Revoke(ctx, stored.ID, newEntity.ID); err != nil {
		return nil, err
	}

	if err := uowTx.Commit(ctx); err != nil {
		return nil, err
	}

	return &RefreshResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		UserID:       stored.UserID,
	}, nil
}

type LogoutInput struct {
	RefreshToken string
}

type LogoutUseCase struct {
	uow    ports.UnitOfWork
	tokens *TokenService
}

func NewLogoutUseCase(uow ports.UnitOfWork, tokens *TokenService) *LogoutUseCase {
	return &LogoutUseCase{uow: uow, tokens: tokens}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, input LogoutInput) error {
	ctx, end := observability.StartUseCaseSpan(ctx, "Logout")
	defer end()

	token := strings.TrimSpace(input.RefreshToken)
	if token == "" {
		return ErrAuthInvalidCredentials
	}
	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = uowTx.Rollback(ctx) }()

	hash := hashFromToken(token)
	stored, err := uowTx.RefreshTokenRepository().GetByHash(ctx, hash)
	if err != nil {
		return err
	}
	if stored.RevokedAt != nil {
		return ErrAuthTokenRevoked
	}
	if err := uowTx.RefreshTokenRepository().Revoke(ctx, stored.ID, ""); err != nil {
		return err
	}
	return uowTx.Commit(ctx)
}

func (uc *LoginUseCase) appendLoginAudit(ctx context.Context, uowTx ports.UnitOfWorkTx, userID string, input LoginInput, success bool, reason string) error {
	audit := &entity.LoginAudit{
		ID:        uuid.NewString(),
		UserID:    userID,
		Email:     input.Email,
		IP:        input.IP,
		Success:   success,
		Reason:    reason,
		CreatedAt: uc.clock.Now().UTC(),
	}
	return uowTx.LoginAuditRepository().Append(ctx, audit)
}

func hashFromToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum[:])
}
