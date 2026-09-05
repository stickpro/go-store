package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/service/mail"
)

// RequestCode is step 1 of passwordless auth: it emails a one-time code to the
// address. Registration and login share this entry point — the account is
// created later, in VerifyCode, only after the code checks out.
//
// Admin addresses are silently ignored (they must use password login); the
// caller always sees the same success so account types can't be probed.
func (s Service) RequestCode(ctx context.Context, email string) error {
	email = normalizeEmail(email)

	if u, err := s.userService.GetUserByEmail(ctx, email); err == nil && u != nil && u.IsAdmin.Bool {
		s.logger.Infow("otp: code request for admin account ignored", "email", email)
		return nil
	}

	code, err := s.otp.issue(ctx, email)
	if err != nil {
		return err
	}

	return s.mailService.Enqueue(ctx, email, mail.LoginCode{
		Code:       code,
		TTLMinutes: int(s.cfg.Auth.OTPCodeTTL.Minutes()),
	})
}

// VerifyCode is step 2: it validates the code, creates the account on first
// login, marks the email verified, and issues an auth token.
func (s Service) VerifyCode(ctx context.Context, email, code string) (*Token, *models.User, error) {
	email = normalizeEmail(email)

	if err := s.otp.redeem(ctx, email, code); err != nil {
		return nil, nil, err
	}

	u, err := s.userService.GetUserByEmail(ctx, email)
	switch {
	case err == nil && u != nil:
		if u.IsAdmin.Bool {
			return nil, nil, ErrUsePasswordLogin
		}
		if u.Banned.Bool {
			return nil, nil, ErrUserBanned
		}
		if !u.EmailVerifiedAt.Valid {
			if mErr := s.userService.MarkEmailVerified(ctx, u.ID); mErr != nil {
				s.logger.Errorw("otp: mark email verified", "error", mErr, "user_id", u.ID)
			}
		}
	case errors.Is(err, ErrUserNotFound):
		u, err = s.userService.CreatePasswordless(ctx, email)
		if err != nil {
			return nil, nil, err
		}
		if mErr := s.mailService.Enqueue(ctx, u.Email, mail.Welcome{Email: u.Email}); mErr != nil {
			s.logger.Errorw("failed to enqueue welcome email", "error", mErr, "email", u.Email)
		}
	default:
		return nil, nil, err
	}

	token, err := s.AuthByUser(ctx, u)
	if err != nil {
		return nil, nil, err
	}
	return token, u, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
