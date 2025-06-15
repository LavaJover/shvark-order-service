package domain

type WalletUsecase interface {
	Freeze(traderID string, orderID string, amount float64) error
	Release(traderID string, orderID string, rewardPercent float64) error
	GetTraderBalance(traderID string) (float64, error)
}