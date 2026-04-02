package contracts

import (
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/constant"
)

type ProductPayload struct {
	ExternalID     string               `json:"external_id"`
	Model          string               `json:"model"`
	Sku            *string              `json:"sku,omitempty"`
	PriceRetail    decimal.Decimal      `json:"price_retail"`
	PriceBusiness  decimal.Decimal      `json:"price_business"`
	PriceWholesale decimal.Decimal      `json:"price_wholesale"`
	StockStatus    constant.StockStatus `json:"stock_status"`
	Quantity       int64                `json:"quantity"`
	IsEnable       bool                 `json:"is_enable"`
}
