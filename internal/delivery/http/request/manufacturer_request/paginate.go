package manufacturer_request

type GetManufacturerWithPagination struct {
	Page     *uint64 `json:"page" query:"page"`
	PageSize *uint64 `json:"page_size" query:"page_size"`
} // @name GetProductWithPaginationRequest
