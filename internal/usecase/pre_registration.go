package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

var ErrPreRegistrationInvalid = errors.New("pre-cadastro invalido")
var ErrPreRegistrationConflict = errors.New("pre-cadastro conflitante")
var ErrPreRegistrationUserExists = errors.New("usuario ja cadastrado")
var ErrPreRegistrationExpired = errors.New("pre-cadastro expirado")
var ErrPreRegistrationBlocked = errors.New("verificacao bloqueada")
var ErrPreRegistrationTokenInvalid = errors.New("token invalido")
var ErrPreRegistrationAlreadyVerified = errors.New("pre-cadastro ja verificado")

const (
	preRegEventEmailVerify = "pre_registration_email_verification"
	preRegEventPhoneVerify = "pre_registration_phone_verification"
)

type PreRegistrationPolicy struct {
	Expiry           time.Duration
	MaxEmailAttempts int
	MaxPhoneAttempts int
	BlockDuration    time.Duration
}

type PreRegistrationInput struct {
	FullName  string
	Email     string
	Phone     string
	IP        string
	UserAgent string
}

type PreRegistrationVerifyInput struct {
	PreRegistrationID string
	Token             string
	IP                string
	Channel           string
}

type PreRegistrationUseCase struct {
	uow    ports.UnitOfWork
	users  repository.UserRepository
	policy PreRegistrationPolicy
	clock  Clock
}

func NewPreRegistrationUseCase(uow ports.UnitOfWork, users repository.UserRepository, policy PreRegistrationPolicy) *PreRegistrationUseCase {
	return &PreRegistrationUseCase{
		uow:    uow,
		users:  users,
		policy: policy,
		clock:  NewRealClock(),
	}
}

func (uc *PreRegistrationUseCase) Start(ctx context.Context, input PreRegistrationInput) (*entity.PreRegistration, error) {
	fullName := strings.TrimSpace(input.FullName)
	email := normalizeEmail(input.Email)
	phone := normalizePhone(input.Phone)
	if fullName == "" || email == "" || phone == "" {
		return nil, ErrPreRegistrationInvalid
	}
	if !isValidEmail(email) || !isValidPhone(phone) {
		return nil, ErrPreRegistrationInvalid
	}

	existingUser, err := uc.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrPreRegistrationUserExists
	}

	now := uc.clock.Now().UTC()
	expiresAt := now.Add(uc.policy.Expiry)

	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = uowTx.Rollback(ctx) }()

	byEmail, err := uowTx.PreRegistrationRepository().GetByEmail(ctx, email)
	if err != nil && err != repository.ErrNotFound {
		return nil, err
	}
	byPhone, err := uowTx.PreRegistrationRepository().GetByPhone(ctx, phone)
	if err != nil && err != repository.ErrNotFound {
		return nil, err
	}
	if byEmail != nil && byPhone != nil && byEmail.ID != byPhone.ID {
		return nil, ErrPreRegistrationConflict
	}

	item := byEmail
	if item == nil {
		item = byPhone
	}

	emailToken, emailHash, err := generateTokenHash(32)
	if err != nil {
		return nil, err
	}
	phoneCode, phoneHash, err := generateNumericCodeHash(6)
	if err != nil {
		return nil, err
	}

	if item == nil {
		item = &entity.PreRegistration{
			ID:            uuid.NewString(),
			FullName:      fullName,
			Email:         email,
			Phone:         phone,
			Status:        entity.PreRegistrationPending,
			EmailStatus:    entity.VerificationPending,
			PhoneStatus:    entity.VerificationPending,
			EmailTokenHash: emailHash,
			PhoneCodeHash:  phoneHash,
			ExpiresAt:      expiresAt,
			CreatedIP:      input.IP,
			UserAgent:      input.UserAgent,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := uowTx.PreRegistrationRepository().Create(ctx, item); err != nil {
			return nil, err
		}
	} else {
		if item.Status == entity.PreRegistrationConverted {
			return nil, ErrPreRegistrationAlreadyVerified
		}
		if item.Status == entity.PreRegistrationInvited || item.Status == entity.PreRegistrationVerified {
			return item, nil
		}
		if isExpired(item.ExpiresAt, now) {
			resetVerification(item)
		}
		if isBlocked(item.EmailBlockedUntil, now) || isBlocked(item.PhoneBlockedUntil, now) {
			return nil, ErrPreRegistrationBlocked
		}
		item.FullName = fullName
		item.Email = email
		item.Phone = phone
		item.EmailTokenHash = emailHash
		item.PhoneCodeHash = phoneHash
		item.ExpiresAt = expiresAt
		item.UpdatedAt = now
		if err := uowTx.PreRegistrationRepository().Update(ctx, item); err != nil {
			return nil, err
		}
	}

	if item.EmailStatus != entity.VerificationVerified {
		_ = uowTx.OutboxRepository().Enqueue(ctx, preRegEventEmailVerify, map[string]any{
			"pre_registration_id": item.ID,
			"email":               item.Email,
			"token":               emailToken,
			"expires_at":          item.ExpiresAt.UTC().Format(time.RFC3339Nano),
		})
	}
	if item.PhoneStatus != entity.VerificationVerified {
		_ = uowTx.OutboxRepository().Enqueue(ctx, preRegEventPhoneVerify, map[string]any{
			"pre_registration_id": item.ID,
			"phone":               item.Phone,
			"code":                phoneCode,
			"expires_at":          item.ExpiresAt.UTC().Format(time.RFC3339Nano),
		})
	}

	if err := uowTx.Commit(ctx); err != nil {
		return nil, err
	}
	return item, nil
}

func (uc *PreRegistrationUseCase) VerifyEmail(ctx context.Context, input PreRegistrationVerifyInput) (*entity.PreRegistration, error) {
	return uc.verify(ctx, input, "EMAIL")
}

func (uc *PreRegistrationUseCase) VerifyPhone(ctx context.Context, input PreRegistrationVerifyInput) (*entity.PreRegistration, error) {
	return uc.verify(ctx, input, "PHONE")
}

func (uc *PreRegistrationUseCase) verify(ctx context.Context, input PreRegistrationVerifyInput, channel string) (*entity.PreRegistration, error) {
	preID := strings.TrimSpace(input.PreRegistrationID)
	if preID == "" || strings.TrimSpace(input.Token) == "" {
		return nil, ErrPreRegistrationInvalid
	}

	now := uc.clock.Now().UTC()
	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = uowTx.Rollback(ctx) }()

	item, err := uowTx.PreRegistrationRepository().GetByID(ctx, preID)
	if err != nil {
		return nil, err
	}

	if isExpired(item.ExpiresAt, now) {
		markExpired(item, now)
		_ = uowTx.PreRegistrationRepository().Update(ctx, item)
		_ = uc.appendAttempt(ctx, uowTx, item, channel, false, "expired", input.IP, now)
		_ = uowTx.Commit(ctx)
		return nil, ErrPreRegistrationExpired
	}

	if channel == "EMAIL" && item.EmailStatus == entity.VerificationVerified {
		return item, ErrPreRegistrationAlreadyVerified
	}
	if channel == "PHONE" && item.PhoneStatus == entity.VerificationVerified {
		return item, ErrPreRegistrationAlreadyVerified
	}

	if channel == "EMAIL" && isBlocked(item.EmailBlockedUntil, now) {
		_ = uc.appendAttempt(ctx, uowTx, item, channel, false, "blocked", input.IP, now)
		_ = uowTx.Commit(ctx)
		return nil, ErrPreRegistrationBlocked
	}
	if channel == "PHONE" && isBlocked(item.PhoneBlockedUntil, now) {
		_ = uc.appendAttempt(ctx, uowTx, item, channel, false, "blocked", input.IP, now)
		_ = uowTx.Commit(ctx)
		return nil, ErrPreRegistrationBlocked
	}

	if !uc.matchesToken(item, input.Token, channel) {
		uc.incrementAttempts(item, channel, now)
		_ = uowTx.PreRegistrationRepository().Update(ctx, item)
		_ = uc.appendAttempt(ctx, uowTx, item, channel, false, "invalid_token", input.IP, now)
		_ = uowTx.Commit(ctx)
		return nil, ErrPreRegistrationTokenInvalid
	}

	uc.markVerified(item, channel, now)
	if item.EmailStatus == entity.VerificationVerified && item.PhoneStatus == entity.VerificationVerified {
		item.Status = entity.PreRegistrationVerified
	}
	item.UpdatedAt = now
	if err := uowTx.PreRegistrationRepository().Update(ctx, item); err != nil {
		return nil, err
	}
	_ = uc.appendAttempt(ctx, uowTx, item, channel, true, "verified", input.IP, now)
	if err := uowTx.Commit(ctx); err != nil {
		return nil, err
	}
	return item, nil
}

func (uc *PreRegistrationUseCase) Lookup(ctx context.Context, email, phone string) (*entity.PreRegistration, error) {
	email = normalizeEmail(email)
	phone = normalizePhone(phone)
	if email == "" && phone == "" {
		return nil, ErrPreRegistrationInvalid
	}
	if email != "" {
		item, err := uc.uow.Begin(ctx)
		if err != nil {
			return nil, err
		}
		defer func() { _ = item.Rollback(ctx) }()
		pre, err := item.PreRegistrationRepository().GetByEmail(ctx, email)
		if err == nil {
			_ = item.Commit(ctx)
			return pre, nil
		}
		if err != repository.ErrNotFound {
			return nil, err
		}
		_ = item.Commit(ctx)
	}
	if phone != "" {
		item, err := uc.uow.Begin(ctx)
		if err != nil {
			return nil, err
		}
		defer func() { _ = item.Rollback(ctx) }()
		pre, err := item.PreRegistrationRepository().GetByPhone(ctx, phone)
		if err == nil {
			_ = item.Commit(ctx)
			return pre, nil
		}
		if err != repository.ErrNotFound {
			return nil, err
		}
		_ = item.Commit(ctx)
	}
	return nil, repository.ErrNotFound
}

func (uc *PreRegistrationUseCase) matchesToken(item *entity.PreRegistration, raw string, channel string) bool {
	hash := hashToken(raw)
	if channel == "EMAIL" {
		return subtleEqual(hash, item.EmailTokenHash)
	}
	return subtleEqual(hash, item.PhoneCodeHash)
}

func (uc *PreRegistrationUseCase) incrementAttempts(item *entity.PreRegistration, channel string, now time.Time) {
	if channel == "EMAIL" {
		item.EmailAttempts++
		if uc.policy.MaxEmailAttempts > 0 && item.EmailAttempts >= uc.policy.MaxEmailAttempts {
			item.EmailStatus = entity.VerificationBlocked
			blockedUntil := now.Add(uc.policy.BlockDuration)
			item.EmailBlockedUntil = &blockedUntil
		}
		return
	}
	item.PhoneAttempts++
	if uc.policy.MaxPhoneAttempts > 0 && item.PhoneAttempts >= uc.policy.MaxPhoneAttempts {
		item.PhoneStatus = entity.VerificationBlocked
		blockedUntil := now.Add(uc.policy.BlockDuration)
		item.PhoneBlockedUntil = &blockedUntil
	}
}

func (uc *PreRegistrationUseCase) markVerified(item *entity.PreRegistration, channel string, now time.Time) {
	if channel == "EMAIL" {
		item.EmailStatus = entity.VerificationVerified
		item.EmailVerifiedAt = &now
		return
	}
	item.PhoneStatus = entity.VerificationVerified
	item.PhoneVerifiedAt = &now
}

func (uc *PreRegistrationUseCase) appendAttempt(ctx context.Context, uowTx ports.UnitOfWorkTx, item *entity.PreRegistration, channel string, success bool, reason string, ip string, now time.Time) error {
	return uowTx.PreRegistrationAttemptRepository().Append(ctx, &entity.PreRegistrationAttempt{
		ID:                uuid.NewString(),
		PreRegistrationID: item.ID,
		Channel:           channel,
		Success:           success,
		Reason:            reason,
		IP:                ip,
		CreatedAt:         now,
	})
}

func isExpired(expiresAt time.Time, now time.Time) bool {
	return now.After(expiresAt)
}

func markExpired(item *entity.PreRegistration, now time.Time) {
	item.Status = entity.PreRegistrationExpired
	if item.EmailStatus != entity.VerificationVerified {
		item.EmailStatus = entity.VerificationExpired
	}
	if item.PhoneStatus != entity.VerificationVerified {
		item.PhoneStatus = entity.VerificationExpired
	}
	item.UpdatedAt = now
}

func resetVerification(item *entity.PreRegistration) {
	item.Status = entity.PreRegistrationPending
	item.EmailStatus = entity.VerificationPending
	item.PhoneStatus = entity.VerificationPending
	item.EmailVerifiedAt = nil
	item.PhoneVerifiedAt = nil
	item.EmailAttempts = 0
	item.PhoneAttempts = 0
	item.EmailBlockedUntil = nil
	item.PhoneBlockedUntil = nil
}

func isBlocked(blockedUntil *time.Time, now time.Time) bool {
	if blockedUntil == nil {
		return false
	}
	return now.Before(*blockedUntil)
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

var phoneDigits = regexp.MustCompile(`[^0-9+]`)

func normalizePhone(value string) string {
	value = strings.TrimSpace(value)
	value = phoneDigits.ReplaceAllString(value, "")
	return value
}

func isValidEmail(value string) bool {
	if value == "" || len(value) > 320 {
		return false
	}
	_, err := mail.ParseAddress(value)
	return err == nil
}

func isValidPhone(value string) bool {
	if !strings.HasPrefix(value, "+") {
		return false
	}
	if len(value) < 10 || len(value) > 18 {
		return false
	}
	for i := 1; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}
	return true
}

func generateTokenHash(size int) (string, string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(buf)
	hash := hashToken(token)
	return token, hash, nil
}

func generateNumericCodeHash(length int) (string, string, error) {
	if length <= 0 {
		return "", "", fmt.Errorf("length invalido")
	}
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		code[i] = byte('0' + int(buf[i])%10)
	}
	value := string(code)
	hash := hashToken(value)
	return value, hash, nil
}

func hashToken(value string) string {
	sum := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", sum[:])
}

func subtleEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
