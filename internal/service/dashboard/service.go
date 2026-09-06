// Package dashboard assembles the admin overview snapshot (order, revenue,
// catalogue and customer counters) in a handful of aggregate SQL queries, so the
// panel no longer fans out to the list endpoints on load.
package dashboard

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/repository/repository_orders"
	"github.com/stickpro/go-store/internal/storage/repository/repository_users"
	"github.com/stickpro/go-store/pkg/logger"
)

// ErrInvalidPeriod is returned when the requested period ends before it starts.
var ErrInvalidPeriod = errors.New("dashboard: 'to' must not be before 'from'")

type IDashboardService interface {
	// Overview returns the admin dashboard snapshot. from/to are optional; when
	// nil the period is the current day in the store timezone.
	Overview(ctx context.Context, from, to *time.Time) (*dto.DashboardDTO, error)
}

type Service struct {
	logger   logger.Logger
	storage  storage.IStorage
	loc      *time.Location
	currency string
}

func New(cfg *config.Config, l logger.Logger, st storage.IStorage) (*Service, error) {
	tz := cfg.Order.Timezone
	if tz == "" {
		tz = "UTC"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("dashboard: load timezone %q: %w", tz, err)
	}

	currency := cfg.Order.Currency
	if currency == "" {
		currency = "RUB"
	}

	return &Service{
		logger:   l,
		storage:  st,
		loc:      loc,
		currency: currency,
	}, nil
}

func (s *Service) Overview(ctx context.Context, from, to *time.Time) (*dto.DashboardDTO, error) {
	now := time.Now().In(s.loc)
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.loc)
	dayEnd := dayStart.AddDate(0, 0, 1)

	periodFrom, periodTo := dayStart, dayEnd
	if from != nil {
		periodFrom = *from
	}
	if to != nil {
		periodTo = *to
	}
	if periodTo.Before(periodFrom) {
		return nil, ErrInvalidPeriod
	}

	// orders.created_at is a naive `timestamp` written with current_timestamp;
	// comparisons assume the database runs in UTC (the standard deployment), so
	// every bound is converted to its UTC wall clock before it hits SQL.
	os, err := s.storage.Orders().DashboardOrderStats(ctx, repository_orders.DashboardOrderStatsParams{
		TodayFrom:  utcTS(dayStart),
		TodayTo:    utcTS(dayEnd),
		PeriodFrom: utcTS(periodFrom),
		PeriodTo:   utcTS(periodTo),
	})
	if err != nil {
		return nil, fmt.Errorf("dashboard: order stats: %w", err)
	}

	cat, err := s.storage.Products().DashboardCatalogStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("dashboard: catalog stats: %w", err)
	}

	cust, err := s.storage.Users().DashboardCustomerStats(ctx, repository_users.DashboardCustomerStatsParams{
		TodayFrom: utcTS(dayStart),
		TodayTo:   utcTS(dayEnd),
	})
	if err != nil {
		return nil, fmt.Errorf("dashboard: customer stats: %w", err)
	}

	aov := decimal.Zero
	if os.PaidOrdersPeriod > 0 {
		aov = os.RevenuePeriod.Div(decimal.NewFromInt(os.PaidOrdersPeriod)).Round(2)
	}

	return &dto.DashboardDTO{
		PeriodFrom: periodFrom,
		PeriodTo:   periodTo,

		OrdersToday: os.Today,
		OrdersTotal: os.Total,
		OrdersByStatus: dto.DashboardOrdersByStatus{
			Pending:    os.StatusPending,
			Paid:       os.StatusPaid,
			Processing: os.StatusProcessing,
			Shipped:    os.StatusShipped,
			Delivered:  os.StatusDelivered,
			Cancelled:  os.StatusCancelled,
			Refunded:   os.StatusRefunded,
		},
		OrdersByPaymentStatus: dto.DashboardOrdersByPaymentStatus{
			Unpaid:   os.PaymentUnpaid,
			Paid:     os.PaymentPaid,
			Refunded: os.PaymentRefunded,
			Failed:   os.PaymentFailed,
		},

		RevenueToday:      os.RevenueToday,
		RevenuePeriod:     os.RevenuePeriod,
		Currency:          s.currency,
		AverageOrderValue: aov,

		Catalog: dto.DashboardCatalog{
			Products:                cat.Products,
			ProductsWithoutVariants: cat.ProductsWithoutVariants,
			Variants:                cat.Variants,
			VariantsOutOfStock:      cat.VariantsOutOfStock,
			Categories:              cat.Categories,
			Collections:             cat.Collections,
		},
		Customers: dto.DashboardCustomers{
			NewToday: cust.NewToday,
			Total:    cust.Total,
		},
	}, nil
}

func utcTS(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t.UTC(), Valid: true}
}
