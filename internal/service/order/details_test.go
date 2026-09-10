package order

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
)

func TestDetailsParamsFromKeepsUntouchedFields(t *testing.T) {
	o := &models.Order{
		ID:            uuid.New(),
		Status:        "new",
		Email:         "old@example.com",
		ShipCityName:  "Москва",
		ShipAddress:   "ул. Старая, 1",
		ShipRecipient: "Пётр",
		Phone:         pgtype.Text{String: "+79990000000", Valid: true},
		ShippingTotal: decOf("0"),
		GrandTotal:    decOf("1000"),
	}

	newAddr := "ул. Новая, 5"
	p := detailsParamsFrom(o, dto.OrderDetailsUpdateDTO{ShipAddress: &newAddr})

	if p.ShipAddress != newAddr {
		t.Fatalf("ShipAddress = %q, want the edited value", p.ShipAddress)
	}
	if p.Email != "old@example.com" || p.ShipCityName != "Москва" || p.ShipRecipient != "Пётр" {
		t.Fatal("untouched fields must keep the order's current values")
	}
	if p.Phone.String != "+79990000000" {
		t.Fatalf("Phone = %q, want kept", p.Phone.String)
	}
}

func TestDetailsParamsFromClearsNothingOnNilComment(t *testing.T) {
	o := &models.Order{Comment: pgtype.Text{String: "keep me", Valid: true}}
	p := detailsParamsFrom(o, dto.OrderDetailsUpdateDTO{})
	if !p.Comment.Valid || p.Comment.String != "keep me" {
		t.Fatalf("Comment = %+v, want unchanged", p.Comment)
	}
}

func TestHasShippingSelection(t *testing.T) {
	code := "cdek_pvz"
	if !(dto.OrderDetailsUpdateDTO{Shipping: dto.ShippingSelection{MethodCode: &code}}).HasShippingSelection() {
		t.Error("a delivery_method_code must trigger a re-quote")
	}
	empty := ""
	if (dto.OrderDetailsUpdateDTO{Shipping: dto.ShippingSelection{MethodCode: &empty}}).HasShippingSelection() {
		t.Error("an empty delivery_method_code must not trigger a re-quote")
	}
	if (dto.OrderDetailsUpdateDTO{}).HasShippingSelection() {
		t.Error("no delivery fields => no re-quote")
	}
}

func TestCheckoutLinesFromItems(t *testing.T) {
	items := []*models.OrderItem{
		{
			VariantID: uuid.NullUUID{UUID: uuid.New(), Valid: true},
			UnitPrice: decOf("500"),
			Quantity:  3,
			WeightKg:  decOf("1.5"),
			LengthCm:  decOf("10"),
		},
	}

	lines := checkoutLinesFromItems(items)
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}
	if !lines[0].unitPrice.Equal(decOf("500")) || lines[0].quantity != 3 {
		t.Fatalf("line pricing not carried over: %+v", lines[0])
	}
	if !lines[0].weightKG.Equal(decOf("1.5")) || !lines[0].lineTotal().Equal(decOf("1500")) {
		t.Fatalf("parcel/total not carried over: weight=%s total=%s", lines[0].weightKG, lines[0].lineTotal())
	}
}
