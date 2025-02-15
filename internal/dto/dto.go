package dto

type AuthRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6,max=50"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type SendCoinRequest struct {
	ToUser string `json:"toUser" validate:"required"`
	Amount int64  `json:"amount" validate:"required,gt=0"`
}

type UserInfo struct {
	Coins       int64           `json:"coins"`
	Inventory   []InventoryItem `json:"inventory"`
	CoinHistory CoinHistory     `json:"coinHistory"`
}

type InventoryItem struct {
	Type     string `json:"type"`
	Quantity int64  `json:"quantity"`
}

type CoinHistory struct {
	Received []CoinTransfer `json:"received"`
	Sent     []CoinTransfer `json:"sent"`
}

type CoinTransfer struct {
	FromUser string `json:"fromUser,omitempty"`
	ToUser   string `json:"toUser,omitempty"`
	Amount   int64  `json:"amount"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
