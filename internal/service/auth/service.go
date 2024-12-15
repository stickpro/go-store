package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/service/user"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/repository/repository_personal_access_tokens"
	"github.com/stickpro/go-store/internal/storage/repository/repository_users"
	"github.com/stickpro/go-store/internal/tools"
	"github.com/stickpro/go-store/internal/tools/hash"
	"github.com/stickpro/go-store/internal/tools/str"
	"github.com/stickpro/go-store/pkg/logger"
	"hash/crc32"
)

type Token struct {
	TokenEntropy string
	CRC32BHash   string
	FullToken    string
}

type IAuth interface {
	RegisterUser(ctx context.Context, dto RegisterDTO) (*models.User, error)
	Auth(ctx context.Context, dto AuthDTO) (*Token, error)
	AuthByUser(ctx context.Context, user *models.User) (*Token, error)
	GetUserByToken(ctx context.Context, hashedToken string) (*models.User, error)
}

type Service struct {
	cfg         *config.Config
	logger      logger.Logger
	userService user.IUser
	storage     storage.IStorage
}

func New(cfg *config.Config, logger logger.Logger, storage storage.IStorage, userService user.IUser) *Service {
	return &Service{
		cfg:         cfg,
		logger:      logger,
		userService: userService,
		storage:     storage,
	}
}

func (s Service) RegisterUser(ctx context.Context, dto RegisterDTO) (*models.User, error) {
	params := &repository_users.CreateParams{
		Email:    dto.Email,
		Password: dto.Password,
		Location: dto.Location,
		Language: dto.Language,
	}
	registeredUser, err := s.userService.StoreUser(ctx, *params)
	if err != nil {
		return nil, err
	}

	return registeredUser, nil
}

func (s Service) Auth(ctx context.Context, dto AuthDTO) (*Token, error) {
	userForAuth, err := s.userService.GetUserByEmail(ctx, dto.Email)
	if err != nil {
		return nil, err
	}

	if userForAuth.Banned.Bool {
		return nil, err
	}

	if !tools.CheckPasswordHash(dto.Password, userForAuth.Password) {
		return nil, err
	}

	token, err := generateTokenString()
	if err != nil {
		return nil, err
	}
	params := repository_personal_access_tokens.CreateParams{
		TokenableType: "user",
		TokenableID:   userForAuth.ID,
		Name:          "AuthToken",
		Token:         hash.SHA256(token.FullToken),
		ExpiresAt:     nil,
	}
	_, err = s.storage.PersonalAccessToken().Create(ctx, params)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (s Service) AuthByUser(ctx context.Context, user *models.User) (*Token, error) {
	token, err := generateTokenString()
	if err != nil {
		return nil, err
	}

	params := repository_personal_access_tokens.CreateParams{
		TokenableType: "user",
		TokenableID:   user.ID,
		Name:          "AuthToken",
		Token:         hash.SHA256(token.FullToken),
		ExpiresAt:     nil,
	}

	_, err = s.storage.PersonalAccessToken().Create(ctx, params)
	if err != nil {
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
