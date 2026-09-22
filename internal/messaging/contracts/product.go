package contracts

import (
	"strings"

	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/constant"
)

// FlexDecimal unmarshals a decimal from either a JSON number or a JSON string
// (the external system sends dimensions/weight as strings, e.g. "8.4").
// An empty string, null or absent value decodes to zero.
type FlexDecimal struct {
	decimal.Decimal
}

func (d *FlexDecimal) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		d.Decimal = decimal.Zero
		return nil
	}
	v, err := decimal.NewFromString(s)
	if err != nil {
		return err
	}
	d.Decimal = v
	return nil
}

type AttributeItem struct {
	Name  string  `json:"name"`
	Slug  string  `json:"slug"`
	Type  string  `json:"type"` // select, number, boolean, text
	Unit  *string `json:"unit,omitempty"`
	Value string  `json:"value"`
}

type ProductPayload struct {
	ExternalID     string               `json:"external_id"`
	Name           string               `json:"name"`
	Model          string               `json:"model"`
	Sku            *string              `json:"sku,omitempty"`
	PriceRetail    decimal.Decimal      `json:"price_retail"`
	PriceBusiness  decimal.Decimal      `json:"price_business"`
	PriceWholesale decimal.Decimal      `json:"price_wholesale"`
	StockStatus    constant.StockStatus `json:"stock_status"`
	Quantity       int64                `json:"quantity"`
	IsEnable       bool                 `json:"is_enable,string"`
	Attributes     []AttributeItem      `json:"attributes,omitempty"`
	ImageMain      *string              `json:"image_main,omitempty"`
	Images         []string             `json:"images,omitempty"`
	Weight         FlexDecimal          `json:"weight"`
	Length         FlexDecimal          `json:"length"`
	Width          FlexDecimal          `json:"width"`
	Height         FlexDecimal          `json:"height"`
}
