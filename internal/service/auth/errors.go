package auth

import "errors"

var (
	// ErrInvalidCredentials is returned for any failed admin password login,
	// regardless of the real cause (unknown email, wrong password, not an admin).
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserBanned is returned when a banned account tries to authenticate.
	ErrUserBanned = errors.New("account is banned")
	// ErrUsePasswordLogin is returned when an admin account is pushed through the
	// passwordless flow.
	ErrUsePasswordLogin = errors.New("admin accounts must log in with a password")

	// ErrCodeInvalid — wrong code (attempt counter was incremented).
	ErrCodeInvalid = errors.New("invalid code")
	// ErrCodeExpired — no active code for this email (expired, never issued, or burned).
	ErrCodeExpired = errors.New("code expired or not found")
	// ErrTooManyAttempts — the code was burned after too many wrong guesses.
	ErrTooManyAttempts = errors.New("too many attempts, request a new code")
	// ErrResendTooSoon — a code was requested again within the resend window.
	ErrResendTooSoon = errors.New("a code was just sent, please wait before requesting another")
	// ErrRateLimited — hourly code-request cap reached for this email.
	ErrRateLimited = errors.New("too many code requests, try again later")
)
