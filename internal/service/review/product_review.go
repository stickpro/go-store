package review

import (
	"context"

	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/service/product"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/internal/storage/repository/repository_product_reviews"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"github.com/stickpro/go-store/pkg/logger"
)

type IProductReviewService interface {
	GetProductReviewsWithPaginate(ctx context.Context, d dto.GetProductReviewsDTO) (*base.FindResponseWithFullPagination[*models.ProductReview], error)
	GetProductReviewByID(ctx context.Context, id uuid.UUID) (*models.ProductReview, error)
	GetProductReviewsByProductID(ctx context.Context, d dto.GetProductReviewsDTO, productID uuid.UUID) (*base.FindResponseWithFullPagination[*models.ProductReview], error)
	CreateProductReview(ctx context.Context, d dto.CreateProductReviewDTO) (*models.ProductReview, error)
	UpdateProductReviewStatus(ctx context.Context, d dto.UpdateProductReviewStatusDTO) error
	DeleteProductReview(ctx context.Context, id uuid.UUID) error
	RestoreProductReview(ctx context.Context, id uuid.UUID) (*models.ProductReview, error)
}

type Service struct {
	cfg            *config.Config
	l              logger.Logger
	storage        storage.IStorage
	productService product.IProductService
}

func New(cfg *config.Config, log logger.Logger, storage storage.IStorage, pService product.IProductService) *Service {
	return &Service{
		cfg:            cfg,
		l:              log,
		storage:        storage,
		productService: pService,
	}
}

func (s *Service) GetProductReviewsWithPaginate(ctx context.Context, d dto.GetProductReviewsDTO) (*base.FindResponseWithFullPagination[*models.ProductReview], error) {
	commonParams := base.NewCommonFindParams()
	if d.PageSize != nil {
		commonParams.PageSize = d.PageSize
	}
	if d.Page != nil {
		commonParams.Page = d.Page
	}

	commonParams.WithDeleted = true

	productReviews, err := s.storage.ProductReviews().GetWithPaginate(ctx, repository_product_reviews.ProductReviewWithPaginationParams{
		CommonFindParams: *commonParams,
	})

	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.l.Debug("error getting product reviews with pagination", parsedErr)
		return nil, parsedErr
	}
	return productReviews, nil
}

func (s *Service) GetProductReviewByID(ctx context.Context, id uuid.UUID) (*models.ProductReview, error) {
	productReview, err := s.storage.ProductReviews().GetByID(ctx, id)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.l.Debug("error getting product review", parsedErr)
		return nil, parsedErr
	}
	return productReview, nil
}

func (s *Service) GetProductReviewsByProductID(ctx context.Context, d dto.GetProductReviewsDTO, productID uuid.UUID) (*base.FindResponseWithFullPagination[*models.ProductReview], error) {
	commonParams := base.NewCommonFindParams()
	if d.PageSize != nil {
		commonParams.PageSize = d.PageSize
	}
	if d.Page != nil {
		commonParams.Page = d.Page
	}

	productReviews, err := s.storage.ProductReviews().GetByProductIDWithPaginate(ctx, repository_product_reviews.ProductReviewWithPaginationParams{
		CommonFindParams: *commonParams,
		ProductID:        uuid.NullUUID{UUID: productID, Valid: true},
	})
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.l.Debug("error getting product reviews with pagination", parsedErr)
		return nil, parsedErr
	}
	return productReviews, nil
}

func (s *Service) CreateProductReview(ctx context.Context, d dto.CreateProductReviewDTO) (*models.ProductReview, error) {
	_, err := s.productService.GetProductByID(ctx, d.ProductID)
	if err != nil {
		return nil, err
	}

	params := repository_product_reviews.CreateParams{
		ProductID: d.ProductID,
		UserID:    d.UserID,
		OrderID:   uuid.NullUUID{},
		Rating:    d.Rating,
		Title:     pgtypeutils.EncodeText(&d.Title),
		Body:      pgtypeutils.EncodeText(&d.Body),
		Status:    constant.ReviewPending.String(),
	}
	productReview, err := s.storage.ProductReviews().Create(ctx, params)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.l.Debug("error create product review", parsedErr)
		return nil, parsedErr
	}
	return productReview, nil
}

func (s *Service) UpdateProductReviewStatus(ctx context.Context, d dto.UpdateProductReviewStatusDTO) error {
	params := repository_product_reviews.UpdateStatusParams{
		ID:     d.ID,
		Status: d.Status.String(),
	}
	err := s.storage.ProductReviews().UpdateStatus(ctx, params)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.l.Debug("error update product review status", parsedErr)
		return parsedErr
	}
	return nil
}

func (s *Service) DeleteProductReview(ctx context.Context, id uuid.UUID) error {
	err := s.storage.ProductReviews().SoftDelete(ctx, id)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.l.Debug("error delete product review", parsedErr)
		return parsedErr
	}
	return nil
}

func (s *Service) RestoreProductReview(ctx context.Context, id uuid.UUID) (*models.ProductReview, error) {
	productReview, err := s.storage.ProductReviews().Restore(ctx, id)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.l.Debug("error restore product review", parsedErr)
		return nil, parsedErr
	}
	return productReview, nil
}
