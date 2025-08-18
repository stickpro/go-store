package attribute_request

type GetAttributeGroupWithPagination struct {
	Page     *uint32 `json:"page" query:"page"`
	PageSize *uint32 `json:"page_size" query:"page_size"`
} // @name GetAttributeGroupWithPagination
