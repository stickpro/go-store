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
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_payment_refunds"
	"github.com/stickpro/go-store/internal/storage/repository/repository_payments"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
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
	// Refund returns money from the order's captured payment. The amount is
	// never chosen by the caller: it is the overpayment (net paid minus the
	// order's current total), or everything left with Full. Calls with the same
	// idempotency key return the original refund instead of sending a new
	// one. ErrRefundPending means the provider was called but its answer is
	// unknown — the money may or may not have moved, so the caller must not
	// retry under a new key.
	Refund(ctx context.Context, d dto.RefundPaymentDTO) (*dto.RefundResultDTO, error)
	// ListRefunds returns every refund attempt made for the order, newest
	// first.
	ListRefunds(ctx context.Context, orderID uuid.UUID) ([]*dto.PaymentRefundDTO, error)
	// VoidPending cancels the order's open payment attempt (one the customer
	// hasn't paid yet) at the provider and marks it failed, so a PaymentURL
	// issued for an old total can't be paid. A no-op when there is none.
	VoidPending(ctx context.Context, orderID uuid.UUID) error
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
//
//nolint:nilnil // a nil payment with no error means "start a new attempt"
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

	var result *models.Payment
	txErr := repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		// Locked so a webhook and an admin refund on the same payment can't
		// interleave their refunded_amount updates.
		existing, err := s.storage.Payments(repository.WithTx(tx)).GetByIDForUpdate(ctx, nr.PaymentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("payment: lookup notification target: %w", err)
		}
		result = existing

		// The provider retries a webhook until it gets back its expected ack,
		// and deliveries can arrive out of order: anything that doesn't move
		// the payment forward is a duplicate or stale, acked without changes.
		from := Status(existing.Status)
		if !canAdvance(from, nr.Status) {
			return nil
		}

		if nr.Status == StatusConfirmed && !nr.Amount.Equal(existing.Amount) {
			return fmt.Errorf("%w: expected %s, provider reported %s", ErrAmountMismatch, existing.Amount, nr.Amount)
		}

		updated, err := s.storage.Payments(repository.WithTx(tx)).UpdateStatus(ctx, repository_payments.UpdateStatusParams{
			ID:                  existing.ID,
			Status:              nr.Status.String(),
			RawLastNotification: nr.RawNotification,
		})
		if err != nil {
			return fmt.Errorf("payment: update from notification: %w", err)
		}
		result = updated

		switch nr.Status {
		case StatusRefunded:
			// A full refund settles everything, including refunds whose
			// provider call timed out and were left pending.
			if err := s.storage.PaymentRefunds(repository.WithTx(tx)).SucceedAllPendingByPaymentID(ctx, existing.ID); err != nil {
				return fmt.Errorf("payment: settle pending refunds: %w", err)
			}
			result, err = s.storage.Payments(repository.WithTx(tx)).ApplyRefund(ctx, repository_payments.ApplyRefundParams{
				ID:             existing.ID,
				RefundedAmount: existing.Amount,
				Status:         StatusRefunded.String(),
			})
			if err != nil {
				return fmt.Errorf("payment: apply full refund: %w", err)
			}
		case StatusPartiallyRefunded:
			// A partial-refund webhook doesn't say which refund it is for or
			// how much it returned. Refunds made through Refund account for
			// themselves; one made outside the shop (e.g. in the acquirer's
			// dashboard) needs its amount reconciled by hand.
			s.logger.Warnw("payment: partial refund notification; refunded amount is only tracked for refunds made through the shop",
				"payment", existing.ID, "refunded_amount", existing.RefundedAmount)
		}
		return nil
	})
	if txErr != nil {
		switch {
		case errors.Is(txErr, ErrNotFound):
			return nil, nr.AckBody, ErrNotFound
		case errors.Is(txErr, ErrAmountMismatch):
			// Nothing a retry would fix — ack, keep the payment unconfirmed.
			return dto.PaymentDTOFromModel(result), nr.AckBody, txErr
		default:
			return nil, "", txErr
		}
	}

	return dto.PaymentDTOFromModel(result), nr.AckBody, nil
}

func (s *Service) Refund(ctx context.Context, d dto.RefundPaymentDTO) (*dto.RefundResultDTO, error) {
	if res, found, err := s.replayRefund(ctx, d); found {
		return res, err
	}

	p, refund, err := s.reserveRefund(ctx, d)
	if err != nil {
		var uniqueErr *pgerror.UniqueConstraintError
		if errors.As(pgerror.ParseError(err), &uniqueErr) {
			// A concurrent request with the same key won the insert.
			res, _, err := s.replayRefund(ctx, d)
			return res, err
		}
		return nil, err
	}

	provider, ok := s.registry.Get(p.Provider)
	if !ok {
		s.resolveRefund(ctx, refund, refundFailed, nil)
		return nil, ErrProviderNotFound
	}

	// From here on the money may move: finish recording the outcome even if
	// the admin's request is cancelled mid-flight.
	ctx = context.WithoutCancel(ctx)

	res, err := provider.Cancel(ctx, CancelRequest{
		PaymentID:         p.ID,
		ProviderPaymentID: *pgtypeutils.DecodeText(p.ProviderPaymentID),
		Amount:            refund.Amount,
	})
	if err != nil {
		if errors.Is(err, ErrProviderRejected) {
			s.resolveRefund(ctx, refund, refundFailed, res.RawResponse)
			return nil, fmt.Errorf("%w: %w", ErrRefundRejected, err)
		}
		// Timeout or network error: the provider may have refunded. The row
		// stays pending (and keeps holding its amount) until a webhook or
		// manual reconciliation settles it.
		s.logger.Errorw("payment: refund outcome unknown, left pending",
			"payment", p.ID, "refund", refund.ID, "amount", refund.Amount, "error", err)
		return nil, fmt.Errorf("%w: %w", ErrRefundPending, err)
	}

	return s.settleRefund(ctx, refund, res)
}

// replayRefund returns the outcome of an earlier refund made under the same
// idempotency key; found is false when the key is new.
func (s *Service) replayRefund(ctx context.Context, d dto.RefundPaymentDTO) (res *dto.RefundResultDTO, found bool, err error) {
	refund, err := s.storage.PaymentRefunds().GetByIdempotencyKey(ctx, d.IdempotencyKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, true, fmt.Errorf("payment: lookup refund by idempotency key: %w", err)
	}

	p, err := s.storage.Payments().GetByID(ctx, refund.PaymentID)
	if err != nil {
		return nil, true, fmt.Errorf("payment: load refunded payment: %w", err)
	}
	if p.OrderID != d.OrderID {
		return nil, true, ErrIdempotencyConflict
	}

	switch refundStatus(refund.Status) {
	case refundSucceeded:
		return &dto.RefundResultDTO{
			Payment: dto.PaymentDTOFromModel(p),
			Refund:  dto.PaymentRefundDTOFromModel(refund),
		}, true, nil
	case refundFailed:
		return nil, true, ErrRefundRejected
	default:
		return nil, true, ErrRefundPending
	}
}

// reserveRefund works out what is owed under lock and records the refund as
// pending before any money moves. The order row is locked first (the same
// lock an item edit takes) so the total the amount is derived from can't
// change underneath it; then the payment, so concurrent refunds see each
// other's pending rows. This is the one place the payment layer reads the
// orders table: the refund amount is defined by the order's total.
func (s *Service) reserveRefund(ctx context.Context, d dto.RefundPaymentDTO) (*models.Payment, *models.PaymentRefund, error) {
	var (
		p      *models.Payment
		refund *models.PaymentRefund
	)

	err := repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		o, err := s.storage.Orders(repository.WithTx(tx)).GetByIDForUpdate(ctx, d.OrderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("payment: lock order for refund: %w", err)
		}

		// A full refund returns everything; otherwise the customer keeps
		// paying for what the order still holds and gets only the excess.
		keep := o.GrandTotal
		if d.Full {
			keep = decimal.Zero
		}

		p, err = s.storage.Payments(repository.WithTx(tx)).GetRefundableByOrderIDForUpdate(ctx, d.OrderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotConfirmed
			}
			return fmt.Errorf("payment: lock for refund: %w", err)
		}
		if pgtypeutils.DecodeText(p.ProviderPaymentID) == nil {
			return fmt.Errorf("payment: captured payment %s has no provider payment id", p.ID)
		}

		pending, err := s.storage.PaymentRefunds(repository.WithTx(tx)).SumPendingByPaymentID(ctx, p.ID)
		if err != nil {
			return fmt.Errorf("payment: sum pending refunds: %w", err)
		}

		amount, err := refundAmount(p.Amount, p.RefundedAmount, pending, keep)
		if err != nil {
			return err
		}

		refund, err = s.storage.PaymentRefunds(repository.WithTx(tx)).Create(ctx, repository_payment_refunds.CreateParams{
			PaymentID:      p.ID,
			Amount:         amount,
			Status:         string(refundPending),
			IdempotencyKey: d.IdempotencyKey,
			Reason:         pgtypeutils.EncodeText(d.Reason),
			Actor:          d.Actor,
		})
		if err != nil {
			return fmt.Errorf("payment: record refund: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return p, refund, nil
}

// refundAmount is what is owed back: what was captured, minus what is
// already refunded or in flight, minus keep (the part the customer still
// pays for — the order total, or zero for a full refund).
func refundAmount(captured, refunded, pending, keep decimal.Decimal) (decimal.Decimal, error) {
	owed := captured.Sub(refunded).Sub(pending).Sub(keep)
	if !owed.IsPositive() {
		if pending.IsPositive() {
			return decimal.Zero, ErrRefundPending
		}
		return decimal.Zero, ErrNothingToRefund
	}
	return owed, nil
}

// settleRefund books a refund the provider accepted onto the payment.
func (s *Service) settleRefund(ctx context.Context, refund *models.PaymentRefund, res CancelResult) (*dto.RefundResultDTO, error) {
	var (
		p       *models.Payment
		settled = refund
	)

	err := repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		var err error
		p, err = s.storage.Payments(repository.WithTx(tx)).GetByIDForUpdate(ctx, refund.PaymentID)
		if err != nil {
			return fmt.Errorf("payment: lock to settle refund: %w", err)
		}

		resolved, err := s.storage.PaymentRefunds(repository.WithTx(tx)).ResolvePending(ctx, repository_payment_refunds.ResolvePendingParams{
			ID:          refund.ID,
			Status:      string(refundSucceeded),
			RawResponse: res.RawResponse,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// A full-refund webhook got here first and already counted it.
				return nil
			}
			return fmt.Errorf("payment: resolve refund: %w", err)
		}
		settled = resolved

		refunded := decimal.Min(p.RefundedAmount.Add(refund.Amount), p.Amount)
		status := StatusPartiallyRefunded
		if refunded.Equal(p.Amount) || res.Status == StatusRefunded {
			status, refunded = StatusRefunded, p.Amount
		}

		p, err = s.storage.Payments(repository.WithTx(tx)).ApplyRefund(ctx, repository_payments.ApplyRefundParams{
			ID:             p.ID,
			RefundedAmount: refunded,
			Status:         status.String(),
		})
		if err != nil {
			return fmt.Errorf("payment: apply refund: %w", err)
		}
		return nil
	})
	if err != nil {
		// The money is gone at the provider but not booked here; the pending
		// row still holds the amount, so nothing can be refunded twice.
		s.logger.Errorw("payment: refund accepted by provider but not recorded",
			"refund", refund.ID, "payment", refund.PaymentID, "amount", refund.Amount, "error", err)
		return nil, err
	}

	return &dto.RefundResultDTO{
		Payment: dto.PaymentDTOFromModel(p),
		Refund:  dto.PaymentRefundDTOFromModel(settled),
	}, nil
}

// resolveRefund marks a pending refund as finished without moving money
// (provider rejected it, or it never reached one). Failure here only leaves
// the row pending, which still blocks a double refund, so it's just logged.
func (s *Service) resolveRefund(ctx context.Context, refund *models.PaymentRefund, status refundStatus, raw []byte) {
	if _, err := s.storage.PaymentRefunds().ResolvePending(ctx, repository_payment_refunds.ResolvePendingParams{
		ID:          refund.ID,
		Status:      string(status),
		RawResponse: raw,
	}); err != nil {
		s.logger.Errorw("payment: resolve refund", "refund", refund.ID, "status", status, "error", err)
	}
}

func (s *Service) ListRefunds(ctx context.Context, orderID uuid.UUID) ([]*dto.PaymentRefundDTO, error) {
	rows, err := s.storage.PaymentRefunds().ListByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("payment: list refunds: %w", err)
	}
	out := make([]*dto.PaymentRefundDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.PaymentRefundDTOFromModel(r))
	}
	return out, nil
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

func (s *Service) VoidPending(ctx context.Context, orderID uuid.UUID) error {
	latest, err := s.storage.Payments().GetLatestByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("payment: lookup for void: %w", err)
	}
	if st := Status(latest.Status); st != StatusNew && st != StatusPending {
		return nil
	}

	if providerPaymentID := pgtypeutils.DecodeText(latest.ProviderPaymentID); providerPaymentID != nil {
		provider, ok := s.registry.Get(latest.Provider)
		if !ok {
			return ErrProviderNotFound
		}
		if _, err := provider.Cancel(ctx, CancelRequest{PaymentID: latest.ID, ProviderPaymentID: *providerPaymentID}); err != nil {
			return fmt.Errorf("payment: provider void: %w", err)
		}
	}

	return repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		// Re-read under lock: the customer may have paid in the meantime, in
		// which case the provider refused the cancel above or the webhook
		// already moved the payment on — either way, leave it alone.
		p, err := s.storage.Payments(repository.WithTx(tx)).GetByIDForUpdate(ctx, latest.ID)
		if err != nil {
			return fmt.Errorf("payment: lock for void: %w", err)
		}
		if !canAdvance(Status(p.Status), StatusFailed) {
			return nil
		}
		if _, err := s.storage.Payments(repository.WithTx(tx)).UpdateStatus(ctx, repository_payments.UpdateStatusParams{
			ID:                  p.ID,
			Status:              StatusFailed.String(),
			RawLastNotification: p.RawLastNotification,
		}); err != nil {
			return fmt.Errorf("payment: mark voided: %w", err)
		}
		return nil
	})
}

var _ IPaymentService = (*Service)(nil)
