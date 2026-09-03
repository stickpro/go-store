package auth

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/service/mail"
	"github.com/stickpro/go-store/internal/service/user"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/repository/repository_personal_access_tokens"
	"github.com/stickpro/go-store/internal/tools"
	"github.com/stickpro/go-store/internal/tools/hash"
	"github.com/stickpro/go-store/internal/tools/str"
	"github.com/stickpro/go-store/pkg/key_value"
	"github.com/stickpro/go-store/pkg/logger"
)

// ErrUserNotFound is re-exported from the user package so callers of this
// package don't need to import both.
var ErrUserNotFound = user.ErrNotFound

type Token struct {
	TokenEntropy string
	CRC32BHash   string
	FullToken    string
}

type IAuthService interface {
	// RequestCode emails a one-time login code (passwordless users only).
	RequestCode(ctx context.Context, email string) error
	// VerifyCode validates a one-time code and returns an auth token, creating
	// the account on first login.
	VerifyCode(ctx context.Context, email, code string) (*Token, error)
	// Auth logs in an admin account with email + password.
	Auth(ctx context.Context, d dto.AuthDTO) (*Token, error)
	AuthByUser(ctx context.Context, user *models.User) (*Token, error)
	GetUserByToken(ctx context.Context, hashedToken string) (*models.User, error)
}

type Service struct {
	cfg         *config.Config
	logger      logger.Logger
	userService user.IUserService
	mailService mail.IMailService
	storage     storage.IStorage
	otp         *codeStore
}

func New(
	cfg *config.Config,
	logger logger.Logger,
	storage storage.IStorage,
	userService user.IUserService,
	mailService mail.IMailService,
	kv key_value.IKeyValue,
) *Service {
	return &Service{
		cfg:         cfg,
		logger:      logger,
		userService: userService,
		mailService: mailService,
		storage:     storage,
		otp:         newCodeStore(kv, cfg.Auth),
	}
}

// Auth authenticates an admin by password. Every failure returns
// ErrInvalidCredentials so the caller can't tell unknown-email from
// wrong-password from not-an-admin.
func (s Service) Auth(ctx context.Context, d dto.AuthDTO) (*Token, error) {
	u, err := s.userService.GetUserByEmail(ctx, normalizeEmail(d.Email))
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !u.IsAdmin.Bool || !u.Password.Valid {
		return nil, ErrInvalidCredentials
	}
	if u.Banned.Bool {
		return nil, ErrUserBanned
	}
	if !tools.CheckPasswordHash(d.Password, u.Password.String) {
		return nil, ErrInvalidCredentials
	}

	return s.AuthByUser(ctx, u)
}

func (s Service) AuthByUser(ctx context.Context, u *models.User) (*Token, error) {
	token, err := generateTokenString()
	if err != nil {
		return nil, err
	}

	params := repository_personal_access_tokens.CreateParams{
		TokenableType: "user",
		TokenableID:   u.ID,
		Name:          "AuthToken",
		Token:         hash.SHA256(token.FullToken),
		ExpiresAt:     nil,
	}

	if _, err := s.storage.PersonalAccessToken().Create(ctx, params); err != nil {
		return nil, err
	}
	return token, nil
}

func (s Service) GetUserByToken(ctx context.Context, hashedToken string) (*models.User, error) {
	token, err := s.storage.PersonalAccessToken().GetByToken(ctx, hashedToken)
	if err != nil {
		return nil, errors.New("token expired")
	}
	u, err := s.userService.GetUserByID(ctx, token.TokenableID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func generateTokenString() (*Token, error) {
	tokenEntropy, err := str.RandomString(40)
	if err != nil {
		return nil, err
	}
	crc32bHash := fmt.Sprintf("%08x", crc32.ChecksumIEEE([]byte(tokenEntropy)))

	fullToken := fmt.Sprintf("%s%s", tokenEntropy, crc32bHash)

	return &Token{
		TokenEntropy: tokenEntropy,
		CRC32BHash:   crc32bHash,
		FullToken:    fullToken,
	}, nil
}
