package constant

type StockStatus string // @name StockStatus

const (
	InStock    StockStatus = "IN_STOCK"
	PreOrder   StockStatus = "PRE_ORDER"
	OutOfStock StockStatus = "OUT_OF_STOCK"
)

func (s StockStatus) String() string {
	return string(s)
}
