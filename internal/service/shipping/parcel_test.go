package shipping

import (
	"testing"

	"github.com/shopspring/decimal"
)

func d(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func TestNewParcel(t *testing.T) {
	def := ParcelDefaults{WeightKG: d("0.5"), LengthCM: d("20"), WidthCM: d("15"), HeightCM: d("10")}

	t.Run("uses actual weight when heavier than volumetric", func(t *testing.T) {
		// 30×20×10 / 5000 = 1.2 kg volumetric; actual 3 kg wins
		p := NewParcel(d("3"), d("30"), d("20"), d("10"), d("0"), def)
		if p.WeightGrams != 3000 {
			t.Fatalf("WeightGrams = %d, want 3000", p.WeightGrams)
		}
		if p.LengthCM != 30 || p.WidthCM != 20 || p.HeightCM != 10 {
			t.Fatalf("dims = %dx%dx%d", p.LengthCM, p.WidthCM, p.HeightCM)
		}
	})

	t.Run("uses volumetric weight when it is heavier", func(t *testing.T) {
		// 60×50×40 / 5000 = 24 kg volumetric; actual 1 kg loses
		p := NewParcel(d("1"), d("60"), d("50"), d("40"), d("0"), def)
		if p.WeightGrams != 24000 {
			t.Fatalf("WeightGrams = %d, want 24000 (volumetric)", p.WeightGrams)
		}
	})

	t.Run("falls back to defaults for zero components", func(t *testing.T) {
		p := NewParcel(d("0"), d("0"), d("0"), d("0"), d("0"), def)
		// default box 20×15×10 → volumetric 600 g, which beats the 500 g default weight
		if p.WeightGrams != 600 || p.LengthCM != 20 || p.WidthCM != 15 || p.HeightCM != 10 {
			t.Fatalf("defaults not applied: %+v", p)
		}
	})

	t.Run("declared value to kopecks", func(t *testing.T) {
		p := NewParcel(d("1"), d("20"), d("15"), d("10"), d("1499.90"), def)
		if p.DeclaredValueKopecks != 149990 {
			t.Fatalf("DeclaredValueKopecks = %d, want 149990", p.DeclaredValueKopecks)
		}
	})
}
