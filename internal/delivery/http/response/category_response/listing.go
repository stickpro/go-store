package category_response

import (
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
)

// BreadcrumbResponse is one node of a breadcrumb trail (root -> ... -> current).
type BreadcrumbResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	MetaTitle *string   `json:"meta_title,omitempty"`
	MetaH1    *string   `json:"meta_h1,omitempty"`
	Depth     int32     `json:"depth"`
} //	@name	BreadcrumbResponse

func NewBreadcrumb(b *dto.BreadcrumbDTO) BreadcrumbResponse {
	return BreadcrumbResponse{
		ID:        b.ID,
		Name:      b.Name,
		Slug:      b.Slug,
		MetaTitle: b.MetaTitle,
		MetaH1:    b.MetaH1,
		Depth:     b.Depth,
	}
}

func NewBreadcrumbs(items []*dto.BreadcrumbDTO) []BreadcrumbResponse {
	out := make([]BreadcrumbResponse, 0, len(items))
	for _, b := range items {
		out = append(out, NewBreadcrumb(b))
	}
	return out
}

// CategoryTreeResponse is a category node with its subtree.
type CategoryTreeResponse struct {
	ID       uuid.UUID              `json:"id"`
	Name     string                 `json:"name"`
	Slug     string                 `json:"slug"`
	Children []CategoryTreeResponse `json:"children,omitempty"`
} //	@name	CategoryTreeResponse

func NewTreeNode(n *dto.CategoryTreeDTO) CategoryTreeResponse {
	node := CategoryTreeResponse{ID: n.ID, Name: n.Name, Slug: n.Slug}
	if len(n.Children) > 0 {
		node.Children = make([]CategoryTreeResponse, 0, len(n.Children))
		for _, ch := range n.Children {
			node.Children = append(node.Children, NewTreeNode(ch))
		}
	}
	return node
}

func NewTree(nodes []*dto.CategoryTreeDTO) []CategoryTreeResponse {
	out := make([]CategoryTreeResponse, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, NewTreeNode(n))
	}
	return out
}

// NewPaginated maps a DB-backed page of categories to the response contract.
func NewPaginated(
	data *base.FindResponseWithFullPagination[*models.Category],
) *base.FindResponseWithFullPagination[CategoryResponse] {
	items := make([]CategoryResponse, 0, len(data.Items))
	for _, c := range data.Items {
		items = append(items, NewFromModel(c))
	}
	return &base.FindResponseWithFullPagination[CategoryResponse]{
		Items:      items,
		Pagination: data.Pagination,
	}
}

// --- category filters ---

type CategoryFilterOptionResponse struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Count int64  `json:"count"`
} //	@name	CategoryFilterOptionResponse

type CategoryPriceRangeResponse struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
} //	@name	CategoryPriceRangeResponse

type CategoryAttributeFilterResponse struct {
	Slug      string                         `json:"slug"`
	Name      string                         `json:"name"`
	Type      string                         `json:"type"`
	Unit      *string                        `json:"unit,omitempty"`
	GroupSlug string                         `json:"group_slug"`
	GroupName string                         `json:"group_name"`
	Options   []CategoryFilterOptionResponse `json:"options,omitempty"`
	Min       *float64                       `json:"min,omitempty"`
	Max       *float64                       `json:"max,omitempty"`
} //	@name	CategoryAttributeFilterResponse

type CategoryFiltersResponse struct {
	Price         *CategoryPriceRangeResponse       `json:"price,omitempty"`
	Manufacturers []CategoryFilterOptionResponse    `json:"manufacturers"`
	StockStatuses []CategoryFilterOptionResponse    `json:"stock_statuses"`
	Attributes    []CategoryAttributeFilterResponse `json:"attributes"`
} //	@name	CategoryFiltersResponse

func newFilterOptions(in []dto.CategoryFilterOptionDTO) []CategoryFilterOptionResponse {
	if in == nil {
		return nil
	}
	out := make([]CategoryFilterOptionResponse, 0, len(in))
	for _, o := range in {
		out = append(out, CategoryFilterOptionResponse{Value: o.Value, Label: o.Label, Count: o.Count})
	}
	return out
}

func NewFilters(f *dto.CategoryFiltersDTO) CategoryFiltersResponse {
	resp := CategoryFiltersResponse{
		Manufacturers: newFilterOptions(f.Manufacturers),
		StockStatuses: newFilterOptions(f.StockStatuses),
	}
	if f.Price != nil {
		resp.Price = &CategoryPriceRangeResponse{Min: f.Price.Min, Max: f.Price.Max}
	}
	resp.Attributes = make([]CategoryAttributeFilterResponse, 0, len(f.Attributes))
	for _, a := range f.Attributes {
		resp.Attributes = append(resp.Attributes, CategoryAttributeFilterResponse{
			Slug:      a.Slug,
			Name:      a.Name,
			Type:      a.Type,
			Unit:      a.Unit,
			GroupSlug: a.GroupSlug,
			GroupName: a.GroupName,
			Options:   newFilterOptions(a.Options),
			Min:       a.Min,
			Max:       a.Max,
		})
	}
	return resp
}
