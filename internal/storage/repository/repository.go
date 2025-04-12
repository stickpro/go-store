package repository

import (
	"github.com/stickpro/go-store/internal/storage/repository/repository_attribute_groups"
	"github.com/stickpro/go-store/internal/storage/repository/repository_attributes"
	"github.com/stickpro/go-store/internal/storage/repository/repository_categories"
	"github.com/stickpro/go-store/internal/storage/repository/repository_manufacturers"
	"github.com/stickpro/go-store/internal/storage/repository/repository_media"
	"github.com/stickpro/go-store/internal/storage/repository/repository_personal_access_tokens"
	"github.com/stickpro/go-store/internal/storage/repository/repository_products"
	"github.com/stickpro/go-store/internal/storage/repository/repository_users"
	"github.com/stickpro/go-store/pkg/database"
	"github.com/stickpro/go-store/pkg/key_value"
)

type IRepository interface {
	Users(opts ...Option) repository_users.Querier
	PersonalAccessToken(opts ...Option) repository_personal_access_tokens.Querier
	Categories(opts ...Option) repository_categories.ICustomQueries
	Products(opts ...Option) repository_products.ICustomQueries
	Media(opts ...Option) repository_media.Querier
	Manufacturers(opts ...Option) repository_manufacturers.ICustomQueries
	AttributeGroups(opts ...Option) repository_attribute_groups.ICustomQueries
	Attributes(opts ...Option) repository_attributes.ICustomQueries
}

type repository struct {
	users               *repository_users.Queries
	personalAccessToken *repository_personal_access_tokens.Queries
	categories          *repository_categories.CustomQueries
	products            *repository_products.CustomQueries
	manufacturer        *repository_manufacturers.CustomQueries
	media               *repository_media.Queries
	attributeGroups     *repository_attribute_groups.CustomQueries
	attributes          *repository_attributes.CustomQueries
}

func InitRepository(psql *database.PostgresClient, keyValue key_value.IKeyValue) IRepository {
	return &repository{
		users:               repository_users.New(psql.DB),
		personalAccessToken: repository_personal_access_tokens.New(psql.DB),
		categories:          repository_categories.NewCustom(psql.DB),
		products:            repository_products.NewCustom(psql.DB),
		manufacturer:        repository_manufacturers.NewCustom(psql.DB),
		media:               repository_media.New(psql.DB),
		attributeGroups:     repository_attribute_groups.NewCustom(psql.DB),
		attributes:          repository_attributes.NewCustom(psql.DB),
	}
}

func (r *repository) Users(opts ...Option) repository_users.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.users.WithTx(options.Tx)
	}
	return r.users
}

func (r *repository) PersonalAccessToken(opts ...Option) repository_personal_access_tokens.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.personalAccessToken.WithTx(options.Tx)
	}
	return r.personalAccessToken
}

func (r *repository) Categories(opts ...Option) repository_categories.ICustomQueries {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.categories.WithTx(options.Tx)
	}
	return r.categories
}

func (r *repository) Products(opts ...Option) repository_products.ICustomQueries {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.products.WithTx(options.Tx)
	}
	return r.products
}

func (r *repository) Manufacturers(opts ...Option) repository_manufacturers.ICustomQueries {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.manufacturer.WithTx(options.Tx)
	}
	return r.manufacturer
}

func (r *repository) Media(opts ...Option) repository_media.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.media.WithTx(options.Tx)
	}
	return r.media
}

func (r *repository) AttributeGroups(opts ...Option) repository_attribute_groups.ICustomQueries {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.attributeGroups.WithTx(options.Tx)
	}
	return r.attributeGroups
}

func (r *repository) Attributes(opts ...Option) repository_attributes.ICustomQueries {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.attributes.WithTx(options.Tx)
	}
	return r.attributes
}
