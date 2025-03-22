package repository

import (
	"github.com/stickpro/go-store/internal/storage/repository/repository_categories"
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
}

type repository struct {
	users               *repository_users.Queries
	personalAccessToken *repository_personal_access_tokens.Queries
	categories          *repository_categories.CustomQueries
	products            *repository_products.CustomQueries
	media               *repository_media.Queries
}

func InitRepository(psql *database.PostgresClient, keyValue key_value.IKeyValue) IRepository {
	return &repository{
		users:               repository_users.New(psql.DB),
		personalAccessToken: repository_personal_access_tokens.New(psql.DB),
		categories:          repository_categories.NewCustom(psql.DB),
		products:            repository_products.NewCustom(psql.DB),
		media:               repository_media.New(psql.DB),
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

func (r *repository) Media(opts ...Option) repository_media.Querier {
	options := parseOptions(opts...)
	if options.Tx != nil {
		return r.media.WithTx(options.Tx)
	}
	return r.media
}
