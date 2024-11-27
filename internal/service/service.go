package service

import (
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/service/auth"
	"github.com/stickpro/go-store/internal/service/user"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/pkg/logger"
)

type Services struct {
	UserService user.IUser
	AuthService auth.IAuth
}

func InitService(
	conf *config.Config,
	logger logger.Logger,
	storage storage.IStorage,
) (*Services, error) {
	userService := user.New(conf, logger, storage)
	authService := auth.New(conf, logger, storage, userService)

	return &Services{
		UserService: userService,
		AuthService: authService,
	}, nil
}
