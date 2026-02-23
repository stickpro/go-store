package contracts

import (
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/constant"
)

type ProductPayload struct {
	ExternalID  string               `json:"external_id"`
	Model       string               `json:"model"`
	Sku         *string              `json:"sku,omitempty"`
	Price       decimal.Decimal      `json:"price"`
	StockStatus constant.StockStatus `json:"stock_status"`
	Quantity    int64                `json:"quantity"`
	IsEnable    bool                 `json:"is_enable"`
}
