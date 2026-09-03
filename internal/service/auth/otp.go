package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/tools/hash"
	"github.com/stickpro/go-store/pkg/key_value"
)

const otpCodeLength = 6

// codeStore owns the lifecycle of email one-time codes: generation, storage,
// verification, and per-email rate limiting. It only talks to the KV store.
type codeStore struct {
	kv  key_value.IKeyValue
	cfg config.AuthConfig
}

func newCodeStore(kv key_value.IKeyValue, cfg config.AuthConfig) *codeStore {
	return &codeStore{kv: kv, cfg: cfg}
}

// record is the JSON blob kept at the code key.
type record struct {
	Hash      string `json:"hash"`
	Attempts  int    `json:"attempts"`
	ExpiresAt int64  `json:"expires_at"` // unix seconds
}

func codeKey(email string) string     { return "otp:code:" + email }
func cooldownKey(email string) string { return "otp:cd:" + email }
func hourlyKey(email string) string   { return "otp:rl:" + email }

// issue generates a fresh code for email and returns the plaintext to be mailed.
// It enforces the resend cooldown and the hourly cap first.
func (c *codeStore) issue(ctx context.Context, email string) (string, error) {
	ok, err := c.kv.SetNX(ctx, cooldownKey(email), "1", c.cfg.OTPResendWindow)
	if err != nil {
		return "", fmt.Errorf("otp: cooldown check: %w", err)
	}
	if !ok {
		return "", ErrResendTooSoon
	}

	n, err := c.kv.Incr(ctx, hourlyKey(email))
	if err != nil {
		return "", fmt.Errorf("otp: hourly counter: %w", err)
	}
	if n == 1 {
		if err := c.kv.Expire(ctx, hourlyKey(email), time.Hour); err != nil {
			return "", fmt.Errorf("otp: hourly ttl: %w", err)
		}
	}
	if int(n) > c.cfg.OTPHourlyLimit {
		return "", ErrRateLimited
	}

	code, err := generateNumericCode(otpCodeLength)
	if err != nil {
		return "", fmt.Errorf("otp: generate code: %w", err)
	}

	rec := record{
		Hash:      hash.SHA256(code),
		Attempts:  0,
		ExpiresAt: time.Now().Add(c.cfg.OTPCodeTTL).Unix(),
	}
	blob, err := json.Marshal(rec)
	if err != nil {
		return "", fmt.Errorf("otp: marshal record: %w", err)
	}
	if err := c.kv.Set(ctx, codeKey(email), string(blob), c.cfg.OTPCodeTTL); err != nil {
		return "", fmt.Errorf("otp: store code: %w", err)
	}

	return code, nil
}

// redeem checks a submitted code. On success the code is consumed. On a wrong
// guess the attempt counter is bumped and the code is burned once it hits the
// limit.
func (c *codeStore) redeem(ctx context.Context, email, code string) error {
	raw, err := c.kv.Get(ctx, codeKey(email))
	if err != nil {
		if errors.Is(err, key_value.ErrEntryNotFound) {
			return ErrCodeExpired
		}
		return fmt.Errorf("otp: load code: %w", err)
	}

	var rec record
	if err := json.Unmarshal(raw.Bytes(), &rec); err != nil {
		return fmt.Errorf("otp: decode record: %w", err)
	}

	if rec.Attempts >= c.cfg.OTPMaxAttempts {
		_ = c.kv.Delete(ctx, codeKey(email))
		return ErrTooManyAttempts
	}

	remaining := time.Until(time.Unix(rec.ExpiresAt, 0))
	if remaining <= 0 {
		_ = c.kv.Delete(ctx, codeKey(email))
		return ErrCodeExpired
	}

	if subtle.ConstantTimeCompare([]byte(rec.Hash), []byte(hash.SHA256(code))) != 1 {
		rec.Attempts++
		if rec.Attempts >= c.cfg.OTPMaxAttempts {
			_ = c.kv.Delete(ctx, codeKey(email))
			return ErrTooManyAttempts
		}
		if blob, mErr := json.Marshal(rec); mErr == nil {
			_ = c.kv.Set(ctx, codeKey(email), string(blob), remaining)
		}
		return ErrCodeInvalid
	}

	_ = c.kv.Delete(ctx, codeKey(email))
	return nil
}

// generateNumericCode returns a cryptographically random decimal string of the
// given length, zero-padded.
func generateNumericCode(length int) (string, error) {
	var b strings.Builder
	b.Grow(length)
	for range length {
		d, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		b.WriteByte(byte('0' + d.Int64()))
	}
	return b.String(), nil
}
