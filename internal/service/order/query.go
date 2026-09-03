package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_orders"
	"github.com/stickpro/go-store/pkg/dbutils"
)

const maxOrderPageSize = 100

func (s *Service) GetByNumber(ctx context.Context, userID uuid.UUID, number int64) (*dto.OrderDTO, error) {
	o, err := s.storage.Orders().GetByNumber(ctx, number)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("order: get by number: %w", err)
	}
	if !o.UserID.Valid || o.UserID.UUID != userID {
		return nil, ErrForbidden
	}
	return s.assemble(ctx, o)
}

func (s *Service) ListForUser(
	ctx context.Context,
	userID uuid.UUID,
	d dto.GetDTO,
) (*base.FindResponseWithFullPagination[*dto.OrderDTO], error) {
	limit, offset, err := dbutils.Pagination(d.Page, d.PageSize, dbutils.WithMaxLimit(maxOrderPageSize))
	if err != nil {
		return nil, err
	}

	nullUser := uuid.NullUUID{UUID: userID, Valid: true}

	total, err := s.storage.Orders().CountByUser(ctx, nullUser)
	if err != nil {
		return nil, fmt.Errorf("order: count by user: %w", err)
	}

	rows, err := s.storage.Orders().ListByUser(ctx, repository_orders.ListByUserParams{
		UserID: nullUser,
		Limit:  int32(limit),  //nolint:gosec // bounded by maxOrderPageSize
		Offset: int32(offset), //nolint:gosec // page arithmetic
	})
	if err != nil {
		return nil, fmt.Errorf("order: list by user: %w", err)
	}

	items, err := s.attachItems(ctx, rows)
	if err != nil {
		return nil, err
	}

	page := uint64(1)
	if d.Page != nil && *d.Page > 0 {
		page = *d.Page
	}
	lastPage := uint64(1)
	if total > 0 {
		lastPage = (uint64(total) + limit - 1) / limit
	}

	return &base.FindResponseWithFullPagination[*dto.OrderDTO]{
		Items: items,
		Pagination: base.FullPagingData{
			Total:    uint64(total), //nolint:gosec // count is non-negative
			PageSize: limit,
			Page:     page,
			LastPage: lastPage,
		},
	}, nil
}

// assemble loads the line items for a single order and maps it to the DTO.
func (s *Service) assemble(ctx context.Context, o *models.Order, opts ...repository.Option) (*dto.OrderDTO, error) {
	its, err := s.storage.OrderItems(opts...).ListByOrderID(ctx, o.ID)
	if err != nil {
		return nil, fmt.Errorf("order: list items: %w", err)
	}
	return dto.OrderDTOFromModel(o, its), nil
}

// attachItems batch-loads line items for a page of orders (one query) and maps
// each order, preserving input order.
func (s *Service) attachItems(ctx context.Context, orders []*models.Order) ([]*dto.OrderDTO, error) {
	if len(orders) == 0 {
		return []*dto.OrderDTO{}, nil
	}

	ids := make([]uuid.UUID, len(orders))
	for i, o := range orders {
		ids[i] = o.ID
	}

	rows, err := s.storage.OrderItems().ListByOrderIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("order: list items batch: %w", err)
	}

	byOrder := make(map[uuid.UUID][]*models.OrderItem, len(orders))
	for _, it := range rows {
		byOrder[it.OrderID] = append(byOrder[it.OrderID], it)
	}

	out := make([]*dto.OrderDTO, len(orders))
	for i, o := range orders {
		out[i] = dto.OrderDTOFromModel(o, byOrder[o.ID])
	}
	return out, nil
}
