// Package invoice issues invoices for "pay by invoice" orders of legal
// entities and renders them to PDF. An invoice is an immutable snapshot
// (seller, buyer, lines, VAT) stored with the row; when the order changes the
// active invoice is voided and a new one with a new number is issued.
package invoice

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"regexp"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_invoices"
)

var (
	// ErrNotFound — the order has no active invoice.
	ErrNotFound = errors.New("invoice: not found")
	// ErrNotConfigured — the seller requisites are not set, so no invoice can
	// be issued.
	ErrNotConfigured = errors.New("invoice: seller requisites are not configured")
	// ErrNotLegal — an invoice was requested for an order without buyer requisites.
	ErrNotLegal = errors.New("invoice: order has no legal-entity buyer")
	// ErrRenderUnavailable — the PDF renderer could not be reached or failed.
	ErrRenderUnavailable = errors.New("invoice: pdf renderer unavailable")
)

type IInvoiceService interface {
	// Ensure makes sure the order has an active invoice matching its current
	// content, issuing one (or voiding the stale one and issuing a new one)
	// when needed. Runs inside the caller's transaction; o must reflect the
	// rows as written by it.
	Ensure(ctx context.Context, tx pgx.Tx, o *dto.OrderDTO) (*dto.InvoiceDTO, error)
	// Void voids the order's active invoice, if any.
	Void(ctx context.Context, tx pgx.Tx, orderID uuid.UUID) error
	// GetActive returns the order's active invoice. tx may be nil.
	GetActive(ctx context.Context, tx pgx.Tx, orderID uuid.UUID) (*dto.InvoiceDTO, error)
	// RenderPDF renders the order's active invoice.
	RenderPDF(ctx context.Context, orderID uuid.UUID) (*dto.InvoicePDF, error)
}

type Service struct {
	cfg      config.InvoiceConfig
	storage  storage.IStorage
	tmpl     *template.Template
	renderer PDFRenderer
	vat      *decimal.Decimal
	now      func() time.Time
}

func New(cfg *config.Config, st storage.IStorage) (*Service, error) {
	loc, err := time.LoadLocation(cfg.Order.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invoice: load timezone %q: %w", cfg.Order.Timezone, err)
	}

	var vat *decimal.Decimal
	if cfg.Invoice.VATRate != "" {
		v, err := decimal.NewFromString(cfg.Invoice.VATRate)
		if err != nil || v.IsNegative() || v.GreaterThan(decimal.NewFromInt(100)) {
			return nil, fmt.Errorf("invoice: vat_rate %q must be a percent between 0 and 100", cfg.Invoice.VATRate)
		}
		vat = &v
	}

	tmpl, err := parseTemplate(cfg.Invoice.TemplatePath, loc)
	if err != nil {
		return nil, err
	}

	return &Service{
		cfg:      cfg.Invoice,
		storage:  st,
		tmpl:     tmpl,
		renderer: newGotenberg(cfg.Invoice.GotenbergURL, cfg.Invoice.RenderTimeout),
		vat:      vat,
		now:      func() time.Time { return time.Now().UTC() },
	}, nil
}

func (s *Service) Ensure(ctx context.Context, tx pgx.Tx, o *dto.OrderDTO) (*dto.InvoiceDTO, error) {
	if !s.cfg.Configured() {
		return nil, ErrNotConfigured
	}
	if o.LegalEntity == nil {
		return nil, ErrNotLegal
	}
	repo := s.storage.Invoices(repository.WithTx(tx))

	now := s.now()
	active, err := repo.GetActiveByOrderID(ctx, o.ID)
	switch {
	case err == nil:
		current, err := decodeSnapshot(active)
		if err != nil {
			return nil, err
		}
		if sameContent(current, buildSnapshot(s.cfg, s.vat, o, active.Number, now, now)) {
			return toDTO(active, current), nil
		}
		if err := repo.VoidActiveByOrderID(ctx, o.ID); err != nil {
			return nil, fmt.Errorf("invoice: void stale: %w", err)
		}
	case !errors.Is(err, pgx.ErrNoRows):
		return nil, fmt.Errorf("invoice: get active: %w", err)
	}

	number, err := repo.NextNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("invoice: next number: %w", err)
	}
	snap := buildSnapshot(s.cfg, s.vat, o, number, now, now.Add(s.cfg.TTL))
	data, err := json.Marshal(snap)
	if err != nil {
		return nil, fmt.Errorf("invoice: marshal snapshot: %w", err)
	}

	created, err := repo.Create(ctx, repository_invoices.CreateParams{
		OrderID:  o.ID,
		Number:   number,
		Amount:   snap.Total,
		Currency: snap.Currency,
		Data:     data,
		IssuedAt: pgtype.Timestamp{Time: snap.IssuedAt, Valid: true},
		DueAt:    pgtype.Timestamp{Time: snap.DueAt, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("invoice: create: %w", err)
	}
	return toDTO(created, snap), nil
}

func (s *Service) Void(ctx context.Context, tx pgx.Tx, orderID uuid.UUID) error {
	if err := s.storage.Invoices(repository.WithTx(tx)).VoidActiveByOrderID(ctx, orderID); err != nil {
		return fmt.Errorf("invoice: void: %w", err)
	}
	return nil
}

func (s *Service) GetActive(ctx context.Context, tx pgx.Tx, orderID uuid.UUID) (*dto.InvoiceDTO, error) {
	inv, snap, err := s.active(ctx, tx, orderID)
	if err != nil {
		return nil, err
	}
	return toDTO(inv, snap), nil
}

func (s *Service) RenderPDF(ctx context.Context, orderID uuid.UUID) (*dto.InvoicePDF, error) {
	inv, snap, err := s.active(ctx, nil, orderID)
	if err != nil {
		return nil, err
	}
	html, err := renderHTML(s.tmpl, snap)
	if err != nil {
		return nil, err
	}
	pdf, err := s.renderer.RenderPDF(ctx, html)
	if err != nil {
		return nil, err
	}
	return &dto.InvoicePDF{FileName: fileName(inv.Number), Content: pdf}, nil
}

func (s *Service) active(ctx context.Context, tx pgx.Tx, orderID uuid.UUID) (*models.Invoice, Snapshot, error) {
	var opts []repository.Option
	if tx != nil {
		opts = append(opts, repository.WithTx(tx))
	}
	inv, err := s.storage.Invoices(opts...).GetActiveByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, Snapshot{}, ErrNotFound
		}
		return nil, Snapshot{}, fmt.Errorf("invoice: get active: %w", err)
	}
	snap, err := decodeSnapshot(inv)
	if err != nil {
		return nil, Snapshot{}, err
	}
	return inv, snap, nil
}

func decodeSnapshot(inv *models.Invoice) (Snapshot, error) {
	var snap Snapshot
	if err := json.Unmarshal(inv.Data, &snap); err != nil {
		return Snapshot{}, fmt.Errorf("invoice: decode snapshot %s: %w", inv.ID, err)
	}
	return snap, nil
}

func toDTO(inv *models.Invoice, snap Snapshot) *dto.InvoiceDTO {
	return &dto.InvoiceDTO{
		ID:       inv.ID,
		OrderID:  inv.OrderID,
		Number:   snap.Number,
		Status:   inv.Status,
		Amount:   inv.Amount,
		Currency: inv.Currency,
		IssuedAt: inv.IssuedAt.Time,
		DueAt:    inv.DueAt.Time,
	}
}

var unsafeFileChars = regexp.MustCompile(`[^0-9A-Za-z_-]+`)

// fileName is an ASCII-only download name, so no header encoding is needed.
func fileName(number int64) string {
	return unsafeFileChars.ReplaceAllString(fmt.Sprintf("invoice-%d", number), "") + ".pdf"
}
