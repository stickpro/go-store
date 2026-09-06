package dashboard_response

import (
	"time"

	"github.com/stickpro/go-store/internal/dto"
)

// DashboardResponse is the admin overview payload. Money fields are decimal
// strings with two fraction digits; period bounds are RFC3339.
type DashboardResponse struct {
	Period            PeriodResponse    `json:"period"`
	Orders            OrdersResponse    `json:"orders"`
	Revenue           RevenueResponse   `json:"revenue"`
	AverageOrderValue string            `json:"average_order_value"`
	Catalog           CatalogResponse   `json:"catalog"`
	Customers         CustomersResponse `json:"customers"`
} //	@name	DashboardResponse

type PeriodResponse struct {
	From string `json:"from"`
	To   string `json:"to"`
} //	@name	DashboardPeriod

type OrdersResponse struct {
	Today           int64                         `json:"today"`
	Total           int64                         `json:"total"`
	ByStatus        OrdersByStatusResponse        `json:"by_status"`
	ByPaymentStatus OrdersByPaymentStatusResponse `json:"by_payment_status"`
} //	@name	DashboardOrders

type OrdersByStatusResponse struct {
	Pending    int64 `json:"pending"`
	Paid       int64 `json:"paid"`
	Processing int64 `json:"processing"`
	Shipped    int64 `json:"shipped"`
	Delivered  int64 `json:"delivered"`
	Cancelled  int64 `json:"cancelled"`
	Refunded   int64 `json:"refunded"`
} //	@name	DashboardOrdersByStatus

type OrdersByPaymentStatusResponse struct {
	Unpaid   int64 `json:"unpaid"`
	Paid     int64 `json:"paid"`
	Refunded int64 `json:"refunded"`
	Failed   int64 `json:"failed"`
} //	@name	DashboardOrdersByPaymentStatus

type RevenueResponse struct {
	Today    string `json:"today"`
	Period   string `json:"period"`
	Currency string `json:"currency"`
	PaidOnly bool   `json:"paid_only"`
} //	@name	DashboardRevenue

type CatalogResponse struct {
	Products                int64 `json:"products"`
	ProductsWithoutVariants int64 `json:"products_without_variants"`
	Variants                int64 `json:"variants"`
	VariantsOutOfStock      int64 `json:"variants_out_of_stock"`
	Categories              int64 `json:"categories"`
	Collections             int64 `json:"collections"`
} //	@name	DashboardCatalog

type CustomersResponse struct {
	NewToday int64 `json:"new_today"`
	Total    int64 `json:"total"`
} //	@name	DashboardCustomers

func NewFromDTO(d *dto.DashboardDTO) *DashboardResponse {
	return &DashboardResponse{
		Period: PeriodResponse{
			From: d.PeriodFrom.Format(time.RFC3339),
			To:   d.PeriodTo.Format(time.RFC3339),
		},
		Orders: OrdersResponse{
			Today: d.OrdersToday,
			Total: d.OrdersTotal,
			ByStatus: OrdersByStatusResponse{
				Pending:    d.OrdersByStatus.Pending,
				Paid:       d.OrdersByStatus.Paid,
				Processing: d.OrdersByStatus.Processing,
				Shipped:    d.OrdersByStatus.Shipped,
				Delivered:  d.OrdersByStatus.Delivered,
				Cancelled:  d.OrdersByStatus.Cancelled,
				Refunded:   d.OrdersByStatus.Refunded,
			},
			ByPaymentStatus: OrdersByPaymentStatusResponse{
				Unpaid:   d.OrdersByPaymentStatus.Unpaid,
				Paid:     d.OrdersByPaymentStatus.Paid,
				Refunded: d.OrdersByPaymentStatus.Refunded,
				Failed:   d.OrdersByPaymentStatus.Failed,
			},
		},
		Revenue: RevenueResponse{
			Today:    d.RevenueToday.StringFixed(2),
			Period:   d.RevenuePeriod.StringFixed(2),
			Currency: d.Currency,
			PaidOnly: true,
		},
		AverageOrderValue: d.AverageOrderValue.StringFixed(2),
		Catalog: CatalogResponse{
			Products:                d.Catalog.Products,
			ProductsWithoutVariants: d.Catalog.ProductsWithoutVariants,
			Variants:                d.Catalog.Variants,
			VariantsOutOfStock:      d.Catalog.VariantsOutOfStock,
			Categories:              d.Catalog.Categories,
			Collections:             d.Catalog.Collections,
		},
		Customers: CustomersResponse{
			NewToday: d.Customers.NewToday,
			Total:    d.Customers.Total,
		},
	}
}
