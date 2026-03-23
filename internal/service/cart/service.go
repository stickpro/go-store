package cart

import (
	"context"

	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/pkg/key_value"
	"github.com/stickpro/go-store/pkg/logger"
)

type ICartService interface {
	GetCart(ctx context.Context, owner dto.CartOwner) (*dto.CartDTO, error)
	AddItem(ctx context.Context, owner dto.CartOwner, d dto.AddCartItemDTO) (*dto.CartDTO, error)
	UpdateQuantity(ctx context.Context, owner dto.CartOwner, variantID uuid.UUID, qty int64) (*dto.CartDTO, error)
	RemoveItem(ctx context.Context, owner dto.CartOwner, variantID uuid.UUID) (*dto.CartDTO, error)
	ClearCart(ctx context.Context, owner dto.CartOwner) error
	MergeCarts(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) (*dto.CartDTO, error)
}

type Service struct {
	cfg     *config.Config
	logger  logger.Logger
	storage storage.IStorage
	kv      key_value.IKeyValue
}

func New(cfg *config.Config, l logger.Logger, st storage.IStorage, kv key_value.IKeyValue) *Service {
	return &Service{
		cfg:     cfg,
		logger:  l,
		storage: st,
		kv:      kv,
	}
}

func (s Service) GetCart(ctx context.Context, owner dto.CartOwner) (*dto.CartDTO, error) {
	//TODO implement me
	panic("implement me")
}

func (s Service) AddItem(ctx context.Context, owner dto.CartOwner, d dto.AddCartItemDTO) (*dto.CartDTO, error) {
	//TODO implement me
	panic("implement me")
}

func (s Service) UpdateQuantity(ctx context.Context, owner dto.CartOwner, variantID uuid.UUID, qty int64) (*dto.CartDTO, error) {
	//TODO implement me
	panic("implement me")
}

func (s Service) RemoveItem(ctx context.Context, owner dto.CartOwner, variantID uuid.UUID) (*dto.CartDTO, error) {
	//TODO implement me
	panic("implement me")
}

func (s Service) ClearCart(ctx context.Context, owner dto.CartOwner) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) MergeCarts(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) (*dto.CartDTO, error) {
	//TODO implement me
	panic("implement me")
}

func cartKey(owner dto.CartOwner) string {
	if owner.UserID != nil {
		return "cart:user:" + owner.UserID.String()
	}
	return "cart:session:" + owner.SessionID.String()
}
