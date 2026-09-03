package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/pkg/key_value"
)

func testCodeStore() *codeStore {
	return newCodeStore(key_value.NewInMemory(), config.AuthConfig{
		OTPCodeTTL:      10 * time.Minute,
		OTPMaxAttempts:  3,
		OTPResendWindow: time.Minute,
		OTPHourlyLimit:  5,
	})
}

func TestCodeStore_IssueAndRedeem(t *testing.T) {
	cs := testCodeStore()
	ctx := context.Background()

	code, err := cs.issue(ctx, "u@example.com")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if len(code) != otpCodeLength {
		t.Fatalf("code %q length = %d", code, len(code))
	}

	if err := cs.redeem(ctx, "u@example.com", code); err != nil {
		t.Fatalf("redeem valid code: %v", err)
	}
	// consumed -> gone
	if err := cs.redeem(ctx, "u@example.com", code); !errors.Is(err, ErrCodeExpired) {
		t.Fatalf("second redeem = %v, want ErrCodeExpired", err)
	}
}

func TestCodeStore_ResendCooldown(t *testing.T) {
	cs := testCodeStore()
	ctx := context.Background()

	if _, err := cs.issue(ctx, "u@example.com"); err != nil {
		t.Fatalf("first issue: %v", err)
	}
	if _, err := cs.issue(ctx, "u@example.com"); !errors.Is(err, ErrResendTooSoon) {
		t.Fatalf("second issue = %v, want ErrResendTooSoon", err)
	}
}

func TestCodeStore_WrongCodeBurnsAfterMaxAttempts(t *testing.T) {
	cs := testCodeStore()
	ctx := context.Background()

	code, err := cs.issue(ctx, "u@example.com")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	wrong := "000000"
	if wrong == code {
		wrong = "111111"
	}

	// attempts 1..2 -> invalid; attempt 3 -> burned
	if err := cs.redeem(ctx, "u@example.com", wrong); !errors.Is(err, ErrCodeInvalid) {
		t.Fatalf("attempt 1 = %v, want ErrCodeInvalid", err)
	}
	if err := cs.redeem(ctx, "u@example.com", wrong); !errors.Is(err, ErrCodeInvalid) {
		t.Fatalf("attempt 2 = %v, want ErrCodeInvalid", err)
	}
	if err := cs.redeem(ctx, "u@example.com", wrong); !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("attempt 3 = %v, want ErrTooManyAttempts", err)
	}
	// even the correct code no longer works
	if err := cs.redeem(ctx, "u@example.com", code); !errors.Is(err, ErrCodeExpired) {
		t.Fatalf("post-burn redeem = %v, want ErrCodeExpired", err)
	}
}

func TestCodeStore_HourlyLimit(t *testing.T) {
	cs := testCodeStore()
	cs.cfg.OTPResendWindow = time.Nanosecond // disable cooldown for this test
	ctx := context.Background()

	for i := 0; i < cs.cfg.OTPHourlyLimit; i++ {
		time.Sleep(time.Millisecond)
		if _, err := cs.issue(ctx, "u@example.com"); err != nil {
			t.Fatalf("issue %d: %v", i+1, err)
		}
	}
	time.Sleep(time.Millisecond)
	if _, err := cs.issue(ctx, "u@example.com"); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("over limit = %v, want ErrRateLimited", err)
	}
}
