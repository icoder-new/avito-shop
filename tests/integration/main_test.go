package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/icoder-new/avito-shop/internal/dto"
	"github.com/icoder-new/avito-shop/internal/models"
	"github.com/icoder-new/avito-shop/internal/storage"
	"github.com/icoder-new/avito-shop/pkg/hash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"testing"

	"github.com/icoder-new/avito-shop/internal/config"
	"github.com/icoder-new/avito-shop/pkg/logger"
)

var (
	testConfig  *config.Config
	testLogger  *logger.Logger
	testStorage storage.IStorage
)

func TestMain(m *testing.M) {
	cfg, err := config.LoadConfig("../../config/config.yml")
	if err != nil {
		panic(err)
	}
	testConfig = cfg

	log, err := logger.New(cfg.Settings.Logger)
	if err != nil {
		panic(err)
	}
	testLogger = log

	os.Exit(m.Run())
}

func TestBuyMerchScenario(t *testing.T) {
	user1 := createTestUser(t, "buyer", "password123")

	token := authorizeUser(t, user1)

	info := getUserInfo(t, token)
	assert.Equal(t, int64(1000), info.Coins)

	buyItem(t, token, "t-shirt")

	updatedInfo := getUserInfo(t, token)
	assert.Equal(t, int64(920), updatedInfo.Coins)
	assert.Len(t, updatedInfo.Inventory, 1)
	assert.Equal(t, "t-shirt", updatedInfo.Inventory[0].Type)
}

func TestSendCoinsScenario(t *testing.T) {
	sender := createTestUser(t, "sender", "password123")
	receiver := createTestUser(t, "receiver", "password123")

	senderToken := authorizeUser(t, sender)

	senderInfo := getUserInfo(t, senderToken)
	assert.Equal(t, int64(1000), senderInfo.Coins)

	sendCoins(t, senderToken, "receiver", 500)

	updatedSenderInfo := getUserInfo(t, senderToken)
	assert.Equal(t, int64(500), updatedSenderInfo.Coins)

	receiverToken := authorizeUser(t, receiver)
	receiverInfo := getUserInfo(t, receiverToken)
	assert.Equal(t, int64(1500), receiverInfo.Coins)
}

func createTestUser(t *testing.T, username, password string) *models.User {
	hasher := hash.NewHasher(hash.Config{
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
	})

	passwordHash, err := hasher.Hash(password)
	require.NoError(t, err)

	user, err := testStorage.User().CreateUser(username, passwordHash)
	require.NoError(t, err)

	err = testStorage.User().UpdateUserCoins(user.ID, 1000)
	require.NoError(t, err)

	return &user
}

func authorizeUser(t *testing.T, user *models.User) string {
	authReq := dto.AuthRequest{
		Username: user.Username,
		Password: "password123",
	}

	resp, err := http.Post(
		"http://localhost:8080/api/auth",
		"application/json",
		bytes.NewBuffer(mustJSON(authReq)),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var authResp dto.AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	require.NoError(t, err)
	require.NotEmpty(t, authResp.Token)

	return authResp.Token
}

func getUserInfo(t *testing.T, token string) dto.UserInfo {
	req, err := http.NewRequest("GET", "http://localhost:8080/api/info", nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var info dto.UserInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	require.NoError(t, err)

	return info
}

func buyItem(t *testing.T, token, itemName string) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("http://localhost:8080/api/buy/%s", itemName),
		nil,
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func sendCoins(t *testing.T, token, recipient string, amount int64) {
	sendReq := dto.SendCoinRequest{
		ToUser: recipient,
		Amount: amount,
	}

	req, err := http.NewRequest(
		"POST",
		"http://localhost:8080/api/sendCoin",
		bytes.NewBuffer(mustJSON(sendReq)),
	)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func mustJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
