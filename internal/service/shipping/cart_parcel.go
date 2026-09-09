package shipping

import (
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/dto"
)

// ParcelItem is one order/cart line reduced to what packing needs: how many, the
// unit price (for the declared value) and the product's own weight/dimensions
// (kg/cm, zero = "not set").
type ParcelItem struct {
	Quantity int64
	Price    decimal.Decimal
	WeightKG decimal.Decimal
	LengthCM decimal.Decimal
	WidthCM  decimal.Decimal
	HeightCM decimal.Decimal
}

// ParcelFromItems builds a single shipment box for a set of lines: total weight
// is Σ(weight × qty), the box is the largest dimension of any line in each axis
// (a crude but safe single-box assumption), the declared value is Σ(price × qty).
// Missing per-product weight/dimensions fall back to def.
func ParcelFromItems(items []ParcelItem, def ParcelDefaults) Parcel {
	var totalWeight, maxL, maxW, maxH, declared decimal.Decimal

	for _, it := range items {
		qty := decimal.NewFromInt(it.Quantity)
		totalWeight = totalWeight.Add(orDefault(it.WeightKG, def.WeightKG).Mul(qty))
		maxL = decimal.Max(maxL, orDefault(it.LengthCM, def.LengthCM))
		maxW = decimal.Max(maxW, orDefault(it.WidthCM, def.WidthCM))
		maxH = decimal.Max(maxH, orDefault(it.HeightCM, def.HeightCM))
		declared = declared.Add(it.Price.Mul(qty))
	}

	return NewParcel(totalWeight, maxL, maxW, maxH, declared, def)
}

// ParcelFromCart adapts enriched cart lines to ParcelFromItems.
func ParcelFromCart(cart []dto.CartItemsDTO, def ParcelDefaults) Parcel {
	items := make([]ParcelItem, len(cart))
	for i, c := range cart {
		items[i] = ParcelItem{
			Quantity: c.Quantity, Price: c.Price,
			WeightKG: c.WeightKG, LengthCM: c.LengthCM, WidthCM: c.WidthCM, HeightCM: c.HeightCM,
		}
	}
	return ParcelFromItems(items, def)
}
