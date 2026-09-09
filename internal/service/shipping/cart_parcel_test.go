package shipping

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/dto"
)

func TestParcelFromCart(t *testing.T) {
	def := ParcelDefaults{WeightKG: d("0.5"), LengthCM: d("20"), WidthCM: d("15"), HeightCM: d("10")}

	items := []dto.CartItemsDTO{
		{Quantity: 2, Price: d("1000"), WeightKG: d("1.5"), LengthCM: d("30"), WidthCM: d("20"), HeightCM: d("10")},
		{Quantity: 1, Price: d("500"), WeightKG: d("0"), LengthCM: d("40"), WidthCM: d("0"), HeightCM: d("25")},
	}

	p := ParcelFromCart(items, def)

	// weight: 1.5×2 + default 0.5×1 = 3.5 kg = 3500 g;
	// volumetric of the 40×20×25 box = 4 kg = 4000 g → wins
	if p.WeightGrams != 4000 {
		t.Fatalf("WeightGrams = %d, want 4000 (volumetric)", p.WeightGrams)
	}
	// box = max in each axis, missing width falls back to default 15
	if p.LengthCM != 40 || p.WidthCM != 20 || p.HeightCM != 25 {
		t.Fatalf("box = %dx%dx%d, want 40x20x25", p.LengthCM, p.WidthCM, p.HeightCM)
	}
	// declared value = Σ price×qty = 2500 rub = 250000 kopecks
	if p.DeclaredValueKopecks != 250000 {
		t.Fatalf("DeclaredValueKopecks = %d, want 250000", p.DeclaredValueKopecks)
	}
}

func TestParcelFromCartEmpty(t *testing.T) {
	def := ParcelDefaults{WeightKG: d("0.5"), LengthCM: d("20"), WidthCM: d("15"), HeightCM: d("10")}
	p := ParcelFromCart(nil, def)
	if p.LengthCM != 20 || p.WidthCM != 15 || p.HeightCM != 10 {
		t.Fatalf("empty cart should give the default box, got %+v", p)
	}
	if p.DeclaredValueKopecks != 0 {
		t.Fatalf("empty cart declared value should be 0, got %d", p.DeclaredValueKopecks)
	}
	_ = decimal.Zero
}
