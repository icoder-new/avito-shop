package service

import (
	"github.com/icoder-new/avito-shop/internal/config"
	"github.com/icoder-new/avito-shop/internal/storage"
	"github.com/icoder-new/avito-shop/pkg/jwt"
	"github.com/icoder-new/avito-shop/pkg/logger"
)

type IService interface {
	Auth() IAuth
	User() IUser
	Coin() ICoin
	Inventory() IInventory
}

type service struct {
	auth      IAuth
	user      IUser
	coin      ICoin
	inventory IInventory
}

func NewService(cfg *config.Config, log *logger.Logger, storage storage.IStorage, manager *jwt.TokenManager) IService {
	return &service{
		auth:      newAuth(cfg, log, storage, manager),
		user:      newUser(cfg, log, storage),
		coin:      newCoin(cfg, log, storage),
		inventory: newInventory(cfg, log, storage),
	}
}

func (s *service) Auth() IAuth {
	return s.auth
}

func (s *service) User() IUser {
	return s.user
}

func (s *service) Coin() ICoin {
	return s.coin
}

func (s *service) Inventory() IInventory {
	return s.inventory
}
