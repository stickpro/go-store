package dto

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/delivery/http/request/dashboard_request"
)

// DashboardDTO is the admin overview snapshot assembled by the dashboard
// service. Order status / payment buckets and the *Total counters are all-time;
// Today* use the store-timezone day boundary; RevenuePeriod / AverageOrderValue
// use the selected period.
type DashboardDTO struct {
	PeriodFrom time.Time
	PeriodTo   time.Time

	OrdersToday           int64
	OrdersTotal           int64
	OrdersByStatus        DashboardOrdersByStatus
	OrdersByPaymentStatus DashboardOrdersByPaymentStatus

	RevenueToday      decimal.Decimal
	RevenuePeriod     decimal.Decimal
	Currency          string
	AverageOrderValue decimal.Decimal

	Catalog   DashboardCatalog
	Customers DashboardCustomers
}

type DashboardOrdersByStatus struct {
	Pending    int64
	Paid       int64
	Processing int64
	Shipped    int64
	Delivered  int64
	Cancelled  int64
	Refunded   int64
}

type DashboardOrdersByPaymentStatus struct {
	Unpaid   int64
	Paid     int64
	Refunded int64
	Failed   int64
}

type DashboardCatalog struct {
	Products                int64
	ProductsWithoutVariants int64
	Variants                int64
	VariantsOutOfStock      int64
	Categories              int64
	Collections             int64
}

type DashboardCustomers struct {
	NewToday int64
	Total    int64
}

// RequestToDashboardPeriod parses the optional from/to query params. Both must
// be RFC3339 when present; nil means "let the service default to today".
func RequestToDashboardPeriod(req *dashboard_request.OverviewRequest) (from, to *time.Time, err error) {
	from, err = parseRFC3339Ptr(req.From)
	if err != nil {
		return nil, nil, errors.New("from: must be an RFC3339 timestamp")
	}
	to, err = parseRFC3339Ptr(req.To)
	if err != nil {
		return nil, nil, errors.New("to: must be an RFC3339 timestamp")
	}
	return from, to, nil
}
