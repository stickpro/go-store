package repository

import (
	"github.com/stickpro/go-store/internal/storage/repository/repository_personal_access_tokens"
	"github.com/stickpro/go-store/internal/storage/repository/repository_users"
	"github.com/stickpro/go-store/pkg/database"
	"github.com/stickpro/go-store/pkg/key_value"
)

type IRepository interface {
	Users(opts ...Option) repository_users.Querier
	PersonalAccessToken(opts ...Option) repository_personal_access_tokens.Querier
}

type repository struct {
	users               *repository_users.Queries
	personalAccessToken *repository_personal_access_tokens.Queries
}

func InitRepository(psql *database.PostgresClient, keyValue key_value.IKeyValue) IRepository {
	return &repository{
		users:               repository_users.New(psql.DB),
		personalAccessToken: repository_personal_access_tokens.New(psql.DB),
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
