package usecase

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrTokenInvalid = errors.New("token invalido")
var ErrTokenExpired = errors.New("token expirado")

type TokenService struct {
	secret    []byte
	accessTTL time.Duration
}

func NewTokenService(secret string, accessTTL time.Duration) *TokenService {
	return &TokenService{
		secret:    []byte(secret),
		accessTTL: accessTTL,
	}
}

func (t *TokenService) GenerateAccessToken(userID string) (string, time.Time, error) {
	if userID == "" {
		return "", time.Time{}, ErrTokenInvalid
	}
	now := time.Now().UTC()
	exp := now.Add(t.accessTTL)
	header := map[string]any{"alg": "HS256", "typ": "JWT"}
	payload := map[string]any{
		"sub": userID,
		"iat": now.Unix(),
		"exp": exp.Unix(),
		"jti": randomID(),
	}
	token, err := t.sign(header, payload)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, exp, nil
}

func (t *TokenService) ParseAccessToken(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", ErrTokenInvalid
	}
	signingInput := parts[0] + "." + parts[1]
	expected := t.signRaw(signingInput)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return "", ErrTokenInvalid
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", ErrTokenInvalid
	}
	var payload map[string]any
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return "", ErrTokenInvalid
	}
	sub, _ := payload["sub"].(string)
	expFloat, _ := payload["exp"].(float64)
	if sub == "" || expFloat == 0 {
		return "", ErrTokenInvalid
	}
	if time.Now().UTC().After(time.Unix(int64(expFloat), 0)) {
		return "", ErrTokenExpired
	}
	return sub, nil
}

func (t *TokenService) GenerateRefreshToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	hash := sha256.Sum256([]byte(token))
	return token, fmt.Sprintf("%x", hash[:]), nil
}

func (t *TokenService) sign(header, payload map[string]any) (string, error) {
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	h := base64.RawURLEncoding.EncodeToString(headerBytes)
	p := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := h + "." + p
	signature := t.signRaw(signingInput)
	return signingInput + "." + signature, nil
}

func (t *TokenService) signRaw(signingInput string) string {
	mac := hmac.New(sha256.New, t.secret)
	_, _ = mac.Write([]byte(signingInput))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func randomID() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err == nil {
		return base64.RawURLEncoding.EncodeToString(buf)
	}
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
