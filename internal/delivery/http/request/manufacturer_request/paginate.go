package manufacturer_request

type GetManufacturerWithPagination struct {
	Page     *uint32 `json:"page" query:"page"`
	PageSize *uint32 `json:"page_size" query:"page_size"`
} // @name GetProductWithPaginationRequest
