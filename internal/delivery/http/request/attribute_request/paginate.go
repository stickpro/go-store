package attribute_request

type GetAttributeGroupWithPagination struct {
	Page     *uint64 `json:"page" query:"page"`
	PageSize *uint64 `json:"page_size" query:"page_size"`
} // @name GetAttributeGroupWithPagination

type GetAttributeWithPagination struct {
	Page     *uint64 `json:"page" query:"page"`
	PageSize *uint64 `json:"page_size" query:"page_size"`
} // @name GetAttributeWithPagination

type GetAttributeValueWithPagination struct {
	Page     *uint64 `json:"page" query:"page"`
	PageSize *uint64 `json:"page_size" query:"page_size"`
} // @name GetAttributeValueWithPagination

type FindAttributeWithPagination struct {
	Attribute string  `json:"attribute" query:"attribute"`
	Page      *uint64 `json:"page" query:"page"`
	PageSize  *uint64 `json:"page_size" query:"page_size"`
} // @name FindAttributeWithPagination
