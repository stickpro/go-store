package product_request

// GetCategoryProductsRequest holds the fixed (non-dynamic) query params of the storefront
// category listing. Attribute filters are passed as dynamic params and parsed separately:
//   - attr.<slug>=v1,v2      value set for select/text/boolean attributes
//   - attr_min.<slug>=<num>  lower bound for number attributes
//   - attr_max.<slug>=<num>  upper bound for number attributes
type GetCategoryProductsRequest struct {
	Page           *uint64 `json:"page" query:"page"`
	PageSize       *uint64 `json:"page_size" query:"page_size"`
	Sort           string  `json:"sort" query:"sort"` // price_asc | price_desc | new | popular | name
	PriceMin       string  `json:"price_min" query:"price_min"`
	PriceMax       string  `json:"price_max" query:"price_max"`
	ManufacturerID string  `json:"manufacturer_id" query:"manufacturer_id"` // comma-separated UUIDs
	StockStatus    string  `json:"stock_status" query:"stock_status"`       // comma-separated statuses
	Facets         bool    `json:"facets" query:"facets"`
} //	@name	GetCategoryProductsRequest
