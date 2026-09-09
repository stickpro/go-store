package shipping

import (
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/config"
)

// volumetricDivisor converts L×W×H in centimetres to a billable weight in
// kilograms — the industry-standard air divisor used by CDEK and Yandex.
const volumetricDivisor = 5000

// Parcel is a shipment normalised for carrier APIs: weight in grams, dimensions
// in whole centimetres, declared value in kopecks. Both CDEK and Russian Post
// take weight in grams and dimensions in centimetres. WeightGrams is already the
// billable weight — max(actual, volumetric).
type Parcel struct {
	WeightGrams          int
	LengthCM             int
	WidthCM              int
	HeightCM             int
	DeclaredValueKopecks int64
}

// ParcelDefaults is the fallback box for products whose own weight/dimensions
// are zero. Weight in kilograms, dimensions in centimetres (as stored on the
// products table).
type ParcelDefaults struct {
	WeightKG decimal.Decimal
	LengthCM decimal.Decimal
	WidthCM  decimal.Decimal
	HeightCM decimal.Decimal
}

// ParcelDefaultsFromConfig parses the decimal-string fallbacks from config.
// An unparseable value becomes zero (i.e. the component is treated as "unset").
func ParcelDefaultsFromConfig(c config.ShippingConfig) ParcelDefaults {
	return ParcelDefaults{
		WeightKG: decOrZero(c.DefaultWeightKG),
		LengthCM: decOrZero(c.DefaultLengthCM),
		WidthCM:  decOrZero(c.DefaultWidthCM),
		HeightCM: decOrZero(c.DefaultHeightCM),
	}
}

func decOrZero(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	return d
}

// NewParcel builds a carrier-ready Parcel from a shipment's weight/size in
// kg/cm, substituting def for any zero component and setting WeightGrams to the
// greater of the actual and the volumetric weight.
func NewParcel(weightKG, lengthCM, widthCM, heightCM, declaredValue decimal.Decimal, def ParcelDefaults) Parcel {
	weightKG = orDefault(weightKG, def.WeightKG)
	lengthCM = orDefault(lengthCM, def.LengthCM)
	widthCM = orDefault(widthCM, def.WidthCM)
	heightCM = orDefault(heightCM, def.HeightCM)

	actualG := kgToGrams(weightKG)
	volG := volumetricGrams(lengthCM, widthCM, heightCM)

	return Parcel{
		WeightGrams:          max(actualG, volG),
		LengthCM:             roundToInt(lengthCM),
		WidthCM:              roundToInt(widthCM),
		HeightCM:             roundToInt(heightCM),
		DeclaredValueKopecks: declaredValue.Mul(decimal.NewFromInt(100)).Round(0).IntPart(),
	}
}

func orDefault(v, def decimal.Decimal) decimal.Decimal {
	if v.IsPositive() {
		return v
	}
	return def
}

func kgToGrams(kg decimal.Decimal) int {
	return int(kg.Mul(decimal.NewFromInt(1000)).Round(0).IntPart())
}

func roundToInt(d decimal.Decimal) int {
	return int(d.Round(0).IntPart())
}

// volumetricGrams = L×W×H (cm) / 5000 → kg → g.
func volumetricGrams(lengthCM, widthCM, heightCM decimal.Decimal) int {
	vol := lengthCM.Mul(widthCM).Mul(heightCM).Div(decimal.NewFromInt(volumetricDivisor))
	return int(vol.Mul(decimal.NewFromInt(1000)).Round(0).IntPart())
}
