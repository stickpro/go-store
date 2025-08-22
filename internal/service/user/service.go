package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_users"
	"github.com/stickpro/go-store/internal/tools"
	"github.com/stickpro/go-store/pkg/logger"
)

type IUserService interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	StoreUser(ctx context.Context, params repository_users.CreateParams) (*models.User, error)
}

type Service struct {
	cfg     *config.Config
	logger  logger.Logger
	storage storage.IStorage
}

func (s Service) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.storage.Users().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s Service) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.storage.Users().GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s Service) StoreUser(ctx context.Context, params repository_users.CreateParams) (*models.User, error) {
	// Hash the password
	hashPassword, err := tools.HashPassword(params.Password)
	if err != nil {
		return nil, err
	}
	params.Password = hashPassword
	// Validate parameters
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("validate params error: %w", err)
	}

	// Check if user already exists
	existingUser, err := s.GetUserByEmail(ctx, params.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	var user *models.User

	// Start transaction TODO
	err = pgx.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		var createErr error

		// Create the user
		user, createErr = s.storage.Users(repository.WithTx(tx)).Create(ctx, params)
		if createErr != nil {
			return createErr
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("store user error: %w", err)
	}

	return user, nil
}

func New(cfg *config.Config, logger logger.Logger, storage storage.IStorage) *Service {
	return &Service{
		cfg:     cfg,
		logger:  logger,
		storage: storage,
	}
}
