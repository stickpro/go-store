package order_request

import "github.com/shopspring/decimal"

// AdminRefundOrderRequest cancels/refunds an order's latest confirmed
// payment at the provider. Amount is optional; omitted it refunds in full.
type AdminRefundOrderRequest struct {
	Amount *decimal.Decimal `json:"amount" validate:"omitempty,gt=0"`
} //	@name	AdminRefundOrderRequest
