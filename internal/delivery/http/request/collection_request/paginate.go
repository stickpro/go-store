package collection_request

type GetCollectionWithPagination struct {
	Page     *uint32 `json:"page" query:"page"`
	PageSize *uint32 `json:"page_size" query:"page_size"`
} // @name GetCollectionWithPagination
