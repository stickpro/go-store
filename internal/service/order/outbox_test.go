package order

import (
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/dto"
)

// The snapshot sent to the ERP must carry the order as the documented JSON:
// ids, version, money as strings, and each line's ERP product id.
func TestOrderSnapshot(t *testing.T) {
	orderID := uuid.MustParse("8f2f4f0c-3b0e-4f43-9a52-2d1a6b0d9c10")
	ext, phone, provider := "19c9f92b-81de-11f1-b9d9-00155ddf6800", "+79990001122", "cdek"

	o := &dto.OrderDTO{
		ID:            orderID,
		Number:        100042,
		Status:        "pending",
		Source:        "checkout",
		PaymentStatus: "unpaid",
		Currency:      "RUB",
		Email:         "buyer@example.com",
		Phone:         &phone,
		Shipping: dto.OrderShippingDTO{
			CityName:  "Москва",
			Address:   "ул. Ленина, д. 1",
			Recipient: "Иванов Иван",
			Provider:  &provider,
		},
		Items: []dto.OrderItemDTO{{
			ExternalID: &ext,
			Name:       "iPhone 15 Pro",
			UnitPrice:  decimal.RequireFromString("1290.00"),
			Quantity:   2,
			LineTotal:  decimal.RequireFromString("2580.00"),
		}},
		Subtotal:      decimal.RequireFromString("2580.00"),
		ShippingTotal: decimal.RequireFromString("350.00"),
		GrandTotal:    decimal.RequireFromString("2930.00"),
		Version:       3,
		CreatedAt:     time.Date(2026, 10, 11, 12, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(orderSnapshot(o))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got struct {
		OrderID     string `json:"order_id"`
		OrderNumber int64  `json:"order_number"`
		Version     int64  `json:"version"`
		Status      string `json:"status"`
		Phone       string `json:"phone"`
		GrandTotal  string `json:"grand_total"`
		Shipping    struct {
			CityName string `json:"city_name"`
			Provider string `json:"provider"`
		} `json:"shipping"`
		Items []struct {
			ExternalID string `json:"external_id"`
			Quantity   int64  `json:"quantity"`
			LineTotal  string `json:"line_total"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal %s: %v", data, err)
	}

	if got.OrderID != orderID.String() || got.OrderNumber != 100042 || got.Version != 3 {
		t.Errorf("ids: order_id=%q order_number=%d version=%d", got.OrderID, got.OrderNumber, got.Version)
	}
	if got.Status != "pending" || got.Phone != phone {
		t.Errorf("status=%q phone=%q", got.Status, got.Phone)
	}
	if got.GrandTotal != "2930" && got.GrandTotal != "2930.00" {
		t.Errorf("grand_total = %q, want a decimal string for 2930", got.GrandTotal)
	}
	if got.Shipping.CityName != "Москва" || got.Shipping.Provider != "cdek" {
		t.Errorf("shipping = %+v", got.Shipping)
	}
	if len(got.Items) != 1 || got.Items[0].ExternalID != ext || got.Items[0].Quantity != 2 {
		t.Fatalf("items = %+v", got.Items)
	}
}
