package collection_response

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

type CollectionResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Slug        string    `json:"slug"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
} //	@name	CollectionResponse

type ShortProductResponse struct {
	ID        uuid.UUID       `json:"id"`
	ProductID uuid.UUID       `json:"product_id"`
	Name      string          `json:"name"`
	Model     string          `json:"model"`
	Slug      string          `json:"slug"`
	Image     pgtype.Text     `json:"image"`
	Price     decimal.Decimal `json:"price"`
	IsEnable  bool            `json:"is_enable"`
} //	@name	ShortProductResponse

type CollectionResponseWithProducts struct {
	ID          uuid.UUID               `json:"id"`
	Name        string                  `json:"name"`
	Description *string                 `json:"description,omitempty"`
	Slug        string                  `json:"slug"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   *time.Time              `json:"updated_at"`
	Products    []*ShortProductResponse `json:"products"`
} //	@name	CollectionWithProductResponse

func NewFromModel(collection *models.Collection) *CollectionResponse {
	return &CollectionResponse{
		ID:          collection.ID,
		Name:        collection.Name,
		Description: pgtypeutils.DecodeText(collection.Description),
		Slug:        collection.Slug,
		CreatedAt:   collection.CreatedAt.Time,
		UpdatedAt:   collection.UpdatedAt.Time,
	}
}

func NewFromModels(collection []*models.Collection) []*CollectionResponse {
	res := make([]*CollectionResponse, 0, len(collection))
	for _, c := range collection {
		res = append(res, NewFromModel(c))
	}
	return res
}

func NewFromDTO(d *dto.WithProductsCollectionDTO, priceGroup constant.PriceGroup) *CollectionResponseWithProducts {
	products := make([]*ShortProductResponse, 0, len(d.Products))
	for _, p := range d.Products {
		products = append(products, &ShortProductResponse{
			ID:        p.ID,
			ProductID: p.ProductID,
			Name:      p.Name,
			Model:     p.Model,
			Slug:      p.Slug,
			Image:     p.Image,
			Price:     selectPrice(p, priceGroup),
			IsEnable:  p.IsEnable,
		})
	}
	return &CollectionResponseWithProducts{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		Slug:        d.Slug,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
		Products:    products,
	}
}

func selectPrice(p *dto.ShortProductDTO, group constant.PriceGroup) decimal.Decimal {
	switch group {
	case constant.PriceGroupBusiness:
		return p.PriceBusiness
	case constant.PriceGroupWholeSale:
		return p.PriceWholeSale
	default:
		return p.PriceRetail
	}
}
