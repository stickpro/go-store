package attribute

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_attribute_values"
	"github.com/stickpro/go-store/internal/storage/repository/repository_attributes"
	"github.com/stickpro/go-store/internal/storage/repository/repository_product_attribute_values"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

// defaultAttributeGroupID — фиксированный UUID дефолтной группы, созданной миграцией.
var defaultAttributeGroupID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

func (s *Service) RunInTx(ctx context.Context, fn func(...repository.Option) error) error {
	return repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		return fn(repository.WithTx(tx))
	})
}

func (s *Service) SyncAttributesFromKafka(ctx context.Context, productID uuid.UUID, items []dto.AttributeKafkaItem, opts ...repository.Option) error {
	valueIDs := make([]uuid.UUID, 0, len(items))

	for _, item := range items {
		attr, err := s.storage.Attributes(opts...).GetOrCreate(ctx, repository_attributes.GetOrCreateParams{
			AttributeGroupID: uuid.NullUUID{UUID: defaultAttributeGroupID, Valid: true},
			Name:             item.Name,
			Slug:             item.Slug,
			Type:             item.Type,
			Unit:             pgtypeutils.EncodeText(item.Unit),
		})
		if err != nil {
			return fmt.Errorf("upsert attribute %q: %w", item.Slug, pgerror.ParseError(err))
		}

		normalized := strings.ToLower(strings.TrimSpace(item.Value))
		var valueNumeric decimal.NullDecimal
		if item.Type == "number" {
			if d, err := decimal.NewFromString(item.Value); err == nil {
				valueNumeric = decimal.NullDecimal{Decimal: d, Valid: true}
			}
		}

		val, err := s.storage.AttributeValues(opts...).GetOrCreate(ctx, repository_attribute_values.GetOrCreateParams{
			AttributeID:     attr.ID,
			Value:           item.Value,
			ValueNormalized: pgtype.Text{String: normalized, Valid: true},
			ValueNumeric:    valueNumeric,
			DisplayOrder:    pgtype.Int4{Int32: 0, Valid: true},
			IsActive:        pgtype.Bool{Bool: true, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("upsert attribute value %q for %q: %w", item.Value, item.Slug, pgerror.ParseError(err))
		}

		valueIDs = append(valueIDs, val.ID)
	}

	pav := s.storage.ProductAttributeValues(opts...)

	if err := pav.RemoveAll(ctx, productID); err != nil {
		return pgerror.ParseError(err)
	}

	if len(valueIDs) == 0 {
		return nil
	}

	return pav.AddBatch(ctx, repository_product_attribute_values.AddBatchParams{
		ProductID:        productID,
		AttributeValueID: valueIDs,
	})
}
