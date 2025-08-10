package orderdto

import "time"

type CreateOrderInput struct {
	MerchantParams
	PaymentSearchParams
	AdvancedParams
	Type string
	ExpiresAt time.Time
}

type MerchantParams struct {
	MerchantID 		string
	MerchantOrderID string
	ClientID 		string
}

type PaymentSearchParams struct {
	CryptoRate      float64
	AmountFiat 		float64
	AmountCrypto    float64
	Currency 		string
	PaymentSystem 	string
	BankInfo 		BankInfo
}

type BankInfo struct {
	BankCode string
	NspkCode string
}

type AdvancedParams struct {
	Shuffle 	int32
	CallbackUrl string
	Recalculated bool
}

// orderRequest := domain.Order{
// 	MerchantID: r.MerchantId, +
// 	AmountFiat: r.AmountFiat, +
// 	Currency: r.Currency, +
// 	Country: r.Country,
// 	ClientID: r.ClientId, +
// 	Status: domain.StatusPending,
// 	PaymentSystem: r.PaymentSystem, +
// 	MerchantOrderID: r.MerchantOrderId, +
// 	Shuffle: r.Shuffle,
// 	ExpiresAt: r.ExpiresAt.AsTime(),
// 	CallbackURL: r.CallbackUrl, +
// 	BankCode: r.BankCode,
// 	NspkCode: r.NspkCode,
// 	Type: r.Type,
// }

type GetAllOrdersInput struct {
    TraderID          string
    MerchantID        string
    OrderID           string
    MerchantOrderID   string
    Status            string
    BankCode          string
    TimeOpeningStart  time.Time
    TimeOpeningEnd    time.Time
    AmountFiatMin     float64
    AmountFiatMax     float64
    Type              string
    DeviceID          string
    Page              int32
    Limit             int32
    Sort              string
}