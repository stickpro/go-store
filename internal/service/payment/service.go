package payment

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/repository/repository_payments"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"github.com/stickpro/go-store/pkg/logger"
)

// IPaymentService drives one payment attempt end to end: creating it against
// a provider, reconciling that provider's webhooks, and cancelling/refunding
// it. It never changes an order's own status — that's the caller's job
// (typically the HTTP handler, wiring this together with order.IOrderService)
// so that this package stays independent of the order domain.
type IPaymentService interface {
	// Init creates a payments row and starts the attempt at the provider,
	// returning the row with its PaymentURL filled in.
	Init(ctx context.Context, d dto.InitPaymentDTO) (*dto.PaymentDTO, error)
	// HandleNotification verifies and applies one provider webhook. ackBody
	// is what the HTTP handler must write back to the provider so it stops
	// retrying; it is returned even on a duplicate/already-terminal
	// notification.
	HandleNotification(ctx context.Context, providerCode string, raw []byte) (p *dto.PaymentDTO, ackBody string, err error)
	// Cancel refunds the order's latest confirmed payment, in full when
	// amount is nil.
	Cancel(ctx context.Context, orderID uuid.UUID, amount *decimal.Decimal) (*dto.PaymentDTO, error)
	// GetLatestByOrderID returns the most recent payment attempt for an
	// order, e.g. so the storefront can re-show an unfinished PaymentURL.
	GetLatestByOrderID(ctx context.Context, orderID uuid.UUID) (*dto.PaymentDTO, error)
}

type Service struct {
	logger   logger.Logger
	storage  storage.IStorage
	registry *Registry
}

func New(l logger.Logger, st storage.IStorage, registry *Registry) *Service {
	return &Service{
		logger:   l,
		storage:  st,
		registry: registry,
	}
}

func (s *Service) Init(ctx context.Context, d dto.InitPaymentDTO) (*dto.PaymentDTO, error) {
	provider, ok := s.registry.Get(d.Provider)
	if !ok {
		return nil, ErrProviderNotFound
	}

	if reused, err := s.reusablePending(ctx, d); err != nil {
		return nil, err
	} else if reused != nil {
		return dto.PaymentDTOFromModel(reused), nil
	}

	created, err := s.storage.Payments().Create(ctx, repository_payments.CreateParams{
		OrderID:  d.OrderID,
		Provider: d.Provider,
		Status:   StatusNew.String(),
		Amount:   d.Amount,
		Currency: d.Currency,
	})
	if err != nil {
		return nil, fmt.Errorf("payment: create row: %w", err)
	}

	res, err := provider.Init(ctx, InitRequest{
		PaymentID:     created.ID,
		OrderNumber:   d.OrderNumber,
		Amount:        d.Amount,
		Currency:      d.Currency,
		Description:   d.Description,
		CustomerEmail: d.CustomerEmail,
	})
	if err != nil {
		if _, markErr := s.storage.Payments().UpdateStatus(ctx, repository_payments.UpdateStatusParams{
			ID:     created.ID,
			Status: StatusFailed.String(),
		}); markErr != nil {
			s.logger.Errorw("payment: mark failed after init error", "payment", created.ID, "error", markErr)
		}
		return nil, fmt.Errorf("payment: provider init: %w", err)
	}

	status := res.Status
	if status == "" {
		status = StatusPending
	}

	updated, err := s.storage.Payments().UpdateAfterInit(ctx, repository_payments.UpdateAfterInitParams{
		ID:                created.ID,
		ProviderPaymentID: pgtypeutils.EncodeText(&res.ProviderPaymentID),
		PaymentUrl:        pgtypeutils.EncodeText(&res.PaymentURL),
		Status:            status.String(),
		RawInitResponse:   res.RawResponse,
	})
	if err != nil {
		return nil, fmt.Errorf("payment: update after init: %w", err)
	}

	return dto.PaymentDTOFromModel(updated), nil
}

// reusablePending returns the order's latest payment attempt if it's still
// usable in place of starting a new one: same provider, still StatusPending
// (so it has a live PaymentURL) and for the same amount. A double-click on
// "pay", a page refresh, or the customer backing out of T-Bank's form and
// retrying would otherwise each open a fresh attempt at the provider for the
// same order. A changed amount (the order was edited since) falls through to
// a fresh Init instead of handing back a PaymentURL for the wrong amount.
func (s *Service) reusablePending(ctx context.Context, d dto.InitPaymentDTO) (*models.Payment, error) {
	latest, err := s.storage.Payments().GetLatestByOrderID(ctx, d.OrderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("payment: check for existing attempt: %w", err)
	}

	if !isReusablePending(latest, d.Provider, d.Amount) {
		return nil, nil
	}
	return latest, nil
}

// isReusablePending is the pure eligibility check behind reusablePending,
// kept separate so it can be unit-tested without a storage fake: same
// provider, still StatusPending (so it has a live PaymentURL) and for the
// same amount.
func isReusablePending(latest *models.Payment, provider string, amount decimal.Decimal) bool {
	return latest.Provider == provider && Status(latest.Status) == StatusPending && latest.Amount.Equal(amount)
}

func (s *Service) HandleNotification(ctx context.Context, providerCode string, raw []byte) (*dto.PaymentDTO, string, error) {
	provider, ok := s.registry.Get(providerCode)
	if !ok {
		return nil, "", ErrProviderNotFound
	}

	nr, err := provider.HandleNotification(ctx, raw)
	if err != nil {
		return nil, "", err
	}

	existing, err := s.storage.Payments().GetByID(ctx, nr.PaymentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nr.AckBody, ErrNotFound
		}
		return nil, "", fmt.Errorf("payment: lookup notification target: %w", err)
	}

	// The provider retries a webhook until it gets back its expected ack, so
	// a payment already in a terminal status is a duplicate delivery, not
	// new information — ack it without reprocessing.
	if Status(existing.Status).IsTerminal() {
		return dto.PaymentDTOFromModel(existing), nr.AckBody, nil
	}

	updated, err := s.storage.Payments().UpdateStatus(ctx, repository_payments.UpdateStatusParams{
		ID:                  existing.ID,
		Status:              nr.Status.String(),
		RawLastNotification: nr.RawNotification,
	})
	if err != nil {
		return nil, "", fmt.Errorf("payment: update from notification: %w", err)
	}

	return dto.PaymentDTOFromModel(updated), nr.AckBody, nil
}

func (s *Service) Cancel(ctx context.Context, orderID uuid.UUID, amount *decimal.Decimal) (*dto.PaymentDTO, error) {
	latest, err := s.storage.Payments().GetLatestByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("payment: lookup for cancel: %w", err)
	}
	if Status(latest.Status) != StatusConfirmed {
		return nil, ErrNotConfirmed
	}

	provider, ok := s.registry.Get(latest.Provider)
	if !ok {
		return nil, ErrProviderNotFound
	}

	providerPaymentID := pgtypeutils.DecodeText(latest.ProviderPaymentID)
	if providerPaymentID == nil {
		return nil, fmt.Errorf("payment: confirmed payment %s has no provider payment id", latest.ID)
	}

	req := CancelRequest{PaymentID: latest.ID, ProviderPaymentID: *providerPaymentID}
	if amount != nil {
		req.Amount = *amount
	}

	res, err := provider.Cancel(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("payment: provider cancel: %w", err)
	}

	updated, err := s.storage.Payments().UpdateStatus(ctx, repository_payments.UpdateStatusParams{
		ID:                  latest.ID,
		Status:              res.Status.String(),
		RawLastNotification: res.RawResponse,
	})
	if err != nil {
		return nil, fmt.Errorf("payment: update after cancel: %w", err)
	}

	return dto.PaymentDTOFromModel(updated), nil
}

func (s *Service) GetLatestByOrderID(ctx context.Context, orderID uuid.UUID) (*dto.PaymentDTO, error) {
	m, err := s.storage.Payments().GetLatestByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("payment: get latest: %w", err)
	}
	return dto.PaymentDTOFromModel(m), nil
}

var _ IPaymentService = (*Service)(nil)
