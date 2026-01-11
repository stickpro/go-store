package dto

type GetDTO struct {
	Page     *uint64 `json:"page" query:"page"`
	PageSize *uint64 `json:"page_size" query:"page_size"`
}
