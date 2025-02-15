package unit

import (
	"context"
	"github.com/icoder-new/avito-shop/internal/config"
	"github.com/icoder-new/avito-shop/internal/dto"
	"github.com/icoder-new/avito-shop/internal/models"
	"github.com/icoder-new/avito-shop/internal/service"
	"github.com/icoder-new/avito-shop/internal/storage"
	"github.com/icoder-new/avito-shop/internal/storage/postgres"
	"github.com/icoder-new/avito-shop/pkg/jwt"
	"github.com/icoder-new/avito-shop/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var (
	testStorage storage.IStorage
	testService service.IService
)

func setupTestEnv(t *testing.T) {
	cfg, err := config.LoadConfig("../../config/config.yml")
	require.NoError(t, err)

	log, err := logger.New(cfg.Settings.Logger)
	require.NoError(t, err)

	storage, err := postgres.NewStorage(context.Background(), log, cfg.GetDSN(), cfg.Settings.DB)
	require.NoError(t, err)
	testStorage = storage

	jwtManager, err := jwt.NewTokenManager(config.JWTCredentials{
		SecretKey: "test-secret-key",
		ExpiresIn: time.Hour,
	})
	require.NoError(t, err)

	testService = service.NewService(cfg, log, testStorage, jwtManager)
}

func cleanup(t *testing.T) {
	if testStorage != nil {
		testStorage.CloseDB()
	}
}

func TestAuthService(t *testing.T) {
	setupTestEnv(t)
	defer cleanup(t)

	assert := assert.New(t)

	t.Run("register and login", func(t *testing.T) {
		authReq := dto.AuthRequest{
			Username: "testuser1",
			Password: "password123",
		}

		resp, err := testService.Auth().Login(authReq)
		assert.NoError(err)
		assert.NotEmpty(resp.Token)

		resp2, err := testService.Auth().Login(authReq)
		assert.NoError(err)
		assert.NotEmpty(resp2.Token)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		authReq := dto.AuthRequest{
			Username: "testuser1",
			Password: "wrongpassword",
		}

		_, err := testService.Auth().Login(authReq)
		assert.Error(err)
	})
}

func TestCoinService(t *testing.T) {
	setupTestEnv(t)
	defer cleanup(t)

	assert := assert.New(t)

	user1 := createTestUser(t, "sender")
	user2 := createTestUser(t, "receiver")

	t.Run("send coins", func(t *testing.T) {
		sendReq := dto.SendCoinRequest{
			ToUser: "receiver",
			Amount: 500,
		}

		err := testService.Coin().Send(user1.ID, sendReq)
		assert.NoError(err)

		sender, err := testService.User().GetInfo(user1.ID)
		assert.NoError(err)
		assert.Equal(int64(500), sender.Coins)

		receiver, err := testService.User().GetInfo(user2.ID)
		assert.NoError(err)
		assert.Equal(int64(1500), receiver.Coins)
	})

	t.Run("insufficient funds", func(t *testing.T) {
		sendReq := dto.SendCoinRequest{
			ToUser: "receiver",
			Amount: 2000,
		}

		err := testService.Coin().Send(user1.ID, sendReq)
		assert.Error(err)
	})
}

func TestInventoryService(t *testing.T) {
	setupTestEnv(t)
	defer cleanup(t)

	assert := assert.New(t)

	user := createTestUser(t, "shopper")

	t.Run("buy item", func(t *testing.T) {
		err := testService.Inventory().BuyItem(user.ID, "t-shirt")
		assert.NoError(err)

		info, err := testService.User().GetInfo(user.ID)
		assert.NoError(err)
		assert.Equal(int64(920), info.Coins)
		assert.Len(info.Inventory, 1)
		assert.Equal("t-shirt", info.Inventory[0].Type)
	})

	t.Run("buy expensive item with insufficient funds", func(t *testing.T) {
		err := testService.Inventory().BuyItem(user.ID, "pink-hoody") // стоит 500
		assert.Error(err)
	})

	t.Run("buy invalid item", func(t *testing.T) {
		err := testService.Inventory().BuyItem(user.ID, "invalid-item")
		assert.Error(err)
	})
}

func createTestUser(t *testing.T, username string) *models.User {
	authReq := dto.AuthRequest{
		Username: username,
		Password: "password123",
	}

	_, err := testService.Auth().Login(authReq)
	require.NoError(t, err)

	user, err := testStorage.User().GetUserByUsername(username)
	require.NoError(t, err)

	return &user
}
