package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/repository/repository_users"
	"github.com/stickpro/go-store/internal/tools"
	"github.com/stickpro/go-store/pkg/logger"
)

// ErrNotFound is returned by the getters when no user matches.
var ErrNotFound = errors.New("user not found")

// ErrEmailTaken is returned when creating a user whose email already exists.
var ErrEmailTaken = errors.New("user with this email already exists")

type IUserService interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	// CreatePasswordless creates a regular (no-password) user with a verified
	// email. Used on first passwordless login.
	CreatePasswordless(ctx context.Context, email string) (*models.User, error)
	// CreateAdmin creates an admin account with a bcrypt-hashed password.
	CreateAdmin(ctx context.Context, email, plainPassword string) (*models.User, error)
	// SetPassword sets/replaces the bcrypt password hash for an existing account.
	SetPassword(ctx context.Context, email, plainPassword string) error
	// MarkEmailVerified stamps email_verified_at if it is not set yet.
	MarkEmailVerified(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	cfg     *config.Config
	logger  logger.Logger
	storage storage.IStorage
}

func New(cfg *config.Config, logger logger.Logger, storage storage.IStorage) *Service {
	return &Service{
		cfg:     cfg,
		logger:  logger,
		storage: storage,
	}
}

func (s Service) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.storage.Users().GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s Service) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.storage.Users().GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s Service) CreatePasswordless(ctx context.Context, email string) (*models.User, error) {
	return s.create(ctx, repository_users.CreateParams{
		Email:           email,
		EmailVerifiedAt: pgtype.Timestamp{Time: nowUTC(), Valid: true},
		Language:        s.defaultLanguage(),
		IsAdmin:         pgtype.Bool{Bool: false, Valid: true},
		Banned:          pgtype.Bool{Bool: false, Valid: true},
	})
}

func (s Service) CreateAdmin(ctx context.Context, email, plainPassword string) (*models.User, error) {
	hashed, err := tools.HashPassword(plainPassword)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	return s.create(ctx, repository_users.CreateParams{
		Email:           email,
		EmailVerifiedAt: pgtype.Timestamp{Time: nowUTC(), Valid: true},
		Password:        pgtype.Text{String: hashed, Valid: true},
		Language:        s.defaultLanguage(),
		IsAdmin:         pgtype.Bool{Bool: true, Valid: true},
		Banned:          pgtype.Bool{Bool: false, Valid: true},
	})
}

func (s Service) SetPassword(ctx context.Context, email, plainPassword string) error {
	hashed, err := tools.HashPassword(plainPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if _, err := s.GetUserByEmail(ctx, email); err != nil {
		return err
	}
	return s.storage.Users().SetPassword(ctx, repository_users.SetPasswordParams{
		Email:    email,
		Password: pgtype.Text{String: hashed, Valid: true},
	})
}

func (s Service) MarkEmailVerified(ctx context.Context, id uuid.UUID) error {
	return s.storage.Users().MarkEmailVerified(ctx, id)
}

func (s Service) create(ctx context.Context, params repository_users.CreateParams) (*models.User, error) {
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("validate params: %w", err)
	}

	if _, err := s.GetUserByEmail(ctx, params.Email); err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	user, err := s.storage.Users().Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

// defaultLanguage is used for accounts created without an explicit locale
// (passwordless sign-up, admin CLI). The DB column also defaults to 'en'.
func (s Service) defaultLanguage() string { return "en" }

func nowUTC() time.Time { return time.Now().UTC() }
