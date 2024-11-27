package repository

import (
	"github.com/stickpro/go-store/pkg/database"
	"github.com/stickpro/go-store/pkg/key_value"
)

type IRepository interface {
}

type repository struct {
}

func InitRepository(psql *database.PostgresClient, keyValue key_value.IKeyValue) IRepository {
	return &repository{}
}
