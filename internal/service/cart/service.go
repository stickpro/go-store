package cart

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage"
	repository_products "github.com/stickpro/go-store/internal/storage/repository/repository_products"
	"github.com/stickpro/go-store/pkg/key_value"
	"github.com/stickpro/go-store/pkg/logger"
)

const (
	cartUserTTL    = 30 * 24 * time.Hour
	cartSessionTTL = 7 * 24 * time.Hour
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
	cart, err := s.loadCart(ctx, owner)
	if err != nil {
		return nil, fmt.Errorf("load cart: %w", err)
	}
	if len(cart.Items) == 0 {
		return &dto.CartDTO{}, nil
	}

	return s.enrichCart(ctx, cart)
}

func (s Service) AddItem(ctx context.Context, owner dto.CartOwner, d dto.AddCartItemDTO) (*dto.CartDTO, error) {
	cart, err := s.loadCart(ctx, owner)
	if err != nil {
		return nil, fmt.Errorf("load cart: %w", err)
	}

	// If the variant is already in the cart — increase quantity, don't duplicate.
	for i, item := range cart.Items {
		if item.VariantID == d.VariantID {
			cart.Items[i].Quantity += int(d.Quantity)
			if err = s.saveCart(ctx, owner, cart); err != nil {
				return nil, fmt.Errorf("save cart: %w", err)
			}
			return s.enrichCart(ctx, cart)
		}
	}

	cart.Items = append(cart.Items, models.CartItem{
		ProductID: d.ProductID,
		VariantID: d.VariantID,
		Quantity:  int(d.Quantity),
		AddedAt:   time.Now(),
	})

	if err = s.saveCart(ctx, owner, cart); err != nil {
		return nil, fmt.Errorf("save cart: %w", err)
	}
	return s.enrichCart(ctx, cart)
}

func (s Service) UpdateQuantity(ctx context.Context, owner dto.CartOwner, variantID uuid.UUID, qty int64) (*dto.CartDTO, error) {
	cart, err := s.loadCart(ctx, owner)
	if err != nil {
		return nil, fmt.Errorf("load cart: %w", err)
	}

	for i, item := range cart.Items {
		if item.VariantID == variantID {
			cart.Items[i].Quantity = int(qty)
			if err = s.saveCart(ctx, owner, cart); err != nil {
				return nil, fmt.Errorf("save cart: %w", err)
			}
			return s.enrichCart(ctx, cart)
		}
	}

	return nil, fmt.Errorf("variant %s not found in cart", variantID)
}

func (s Service) RemoveItem(ctx context.Context, owner dto.CartOwner, variantID uuid.UUID) (*dto.CartDTO, error) {
	cart, err := s.loadCart(ctx, owner)
	if err != nil {
		return nil, fmt.Errorf("load cart: %w", err)
	}

	for i, item := range cart.Items {
		if item.VariantID == variantID {
			cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
			if err = s.saveCart(ctx, owner, cart); err != nil {
				return nil, fmt.Errorf("save cart: %w", err)
			}
			return s.enrichCart(ctx, cart)
		}
	}

	return nil, fmt.Errorf("variant %s not found in cart", variantID)
}

func (s Service) ClearCart(ctx context.Context, owner dto.CartOwner) error {
	return s.kv.Delete(ctx, cartKey(owner))
}

// MergeCarts merges a guest cart into a user cart after login.
// Items present in both are combined (quantities are summed).
// The session cart is deleted afterwards.
func (s Service) MergeCarts(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) (*dto.CartDTO, error) {
	sessionOwner := dto.CartOwner{SessionID: &sessionID}
	userOwner := dto.CartOwner{UserID: &userID}

	sessionCart, err := s.loadCart(ctx, sessionOwner)
	if err != nil {
		return nil, fmt.Errorf("load session cart: %w", err)
	}
	if len(sessionCart.Items) == 0 {
		return s.GetCart(ctx, userOwner)
	}

	userCart, err := s.loadCart(ctx, userOwner)
	if err != nil {
		return nil, fmt.Errorf("load user cart: %w", err)
	}

	// Index user cart items for O(1) lookup during merge.
	userItemIdx := make(map[uuid.UUID]int, len(userCart.Items))
	for i, item := range userCart.Items {
		userItemIdx[item.VariantID] = i
	}

	for _, sessionItem := range sessionCart.Items {
		if idx, exists := userItemIdx[sessionItem.VariantID]; exists {
			userCart.Items[idx].Quantity += sessionItem.Quantity
		} else {
			userCart.Items = append(userCart.Items, sessionItem)
		}
	}

	if err = s.saveCart(ctx, userOwner, userCart); err != nil {
		return nil, fmt.Errorf("save merged cart: %w", err)
	}
	if err = s.kv.Delete(ctx, cartKey(sessionOwner)); err != nil {
		return nil, fmt.Errorf("delete session cart: %w", err)
	}

	return s.enrichCart(ctx, userCart)
}

// loadCart reads the raw cart (only IDs + quantities) from Redis.
// Returns an empty cart if the key does not exist yet.
func (s Service) loadCart(ctx context.Context, owner dto.CartOwner) (*models.Cart, error) {
	data, err := s.kv.Get(ctx, cartKey(owner))
	if err != nil {
		if errors.Is(err, key_value.ErrEntryNotFound) {
			return &models.Cart{}, nil
		}
		return nil, err
	}

	var cart models.Cart
	if err = json.Unmarshal(data.Bytes(), &cart); err != nil {
		return nil, fmt.Errorf("unmarshal cart: %w", err)
	}
	return &cart, nil
}

// enrichCart fetches actual prices, names and availability from the DB
// and maps the raw cart into CartDTO.
func (s Service) enrichCart(ctx context.Context, cart *models.Cart) (*dto.CartDTO, error) {
	variantIDs := make([]uuid.UUID, len(cart.Items))
	for i, item := range cart.Items {
		variantIDs[i] = item.VariantID
	}

	rows, err := s.storage.Products().GetCartItemsByVariantIDs(ctx, variantIDs)
	if err != nil {
		return nil, fmt.Errorf("get cart products: %w", err)
	}

	// Index DB rows by variant_id for O(1) lookup.
	rowByVariant := make(map[uuid.UUID]*repository_products.GetCartItemsByVariantIDsRow, len(rows))
	for _, row := range rows {
		rowByVariant[row.VariantID] = row
	}

	result := &dto.CartDTO{
		Items: make([]dto.CartItemsDTO, 0, len(cart.Items)),
	}

	var totalPrice decimal.Decimal
	for _, item := range cart.Items {
		row, ok := rowByVariant[item.VariantID]
		if !ok {
			// Variant was deleted from DB — skip silently.
			continue
		}

		available := row.ProductEnabled && row.VariantEnabled && row.MaxQuantity > 0
		qty := min(item.Quantity, int(row.MaxQuantity))

		imageURL := ""
		if row.Image.Valid {
			imageURL = row.Image.String
		}

		result.Items = append(result.Items, dto.CartItemsDTO{
			ProductID:   row.ProductID,
			VariantID:   row.VariantID,
			Name:        row.Name,
			Slug:        row.Slug,
			ImageURL:    imageURL,
			Price:       row.Price,
			Quantity:    int64(qty),
			MaxQuantity: row.MaxQuantity,
			Available:   available,
		})

		totalPrice = totalPrice.Add(row.Price.Mul(decimal.NewFromInt(int64(qty))))
	}

	result.TotalPrice = totalPrice
	return result, nil
}

// saveCart serialises the cart and writes it to Redis with the appropriate TTL.
func (s Service) saveCart(ctx context.Context, owner dto.CartOwner, cart *models.Cart) error {
	data, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("marshal cart: %w", err)
	}
	return s.kv.Set(ctx, cartKey(owner), string(data), cartTTL(owner))
}

func cartKey(owner dto.CartOwner) string {
	if owner.UserID != nil {
		return "cart:user:" + owner.UserID.String()
	}
	return "cart:session:" + owner.SessionID.String()
}

func cartTTL(owner dto.CartOwner) time.Duration {
	if owner.UserID != nil {
		return cartUserTTL
	}
	return cartSessionTTL
}
