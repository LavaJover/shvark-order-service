package orderdto

import "github.com/LavaJover/shvark-order-service/internal/domain"

type OrderOutput struct {
	Order domain.Order
	BankDetail domain.BankDetail
}


// Order: &orderpb.Order{
// 	OrderId: savedOrder.ID,
// 	Status: string(savedOrder.Status),
// 	Type: savedOrder.Type,
// 	BankDetail: &orderpb.BankDetail{
// 		BankDetailId: savedOrder.BankDetail.ID,
// 		TraderId: savedOrder.BankDetail.TraderID,
// 		Currency: savedOrder.BankDetail.Currency,
// 		Country: savedOrder.BankDetail.Country, 
// 		MinAmount: float64(savedOrder.BankDetail.MinAmount),
// 		MaxAmount: float64(savedOrder.BankDetail.MaxAmount),
// 		BankName: savedOrder.BankDetail.BankName,
// 		PaymentSystem: savedOrder.BankDetail.PaymentSystem,
// 		Enabled: savedOrder.BankDetail.Enabled,
// 		Delay: durationpb.New(savedOrder.BankDetail.Delay),
// 		Owner: savedOrder.BankDetail.Owner,
// 		CardNumber: savedOrder.BankDetail.CardNumber,
// 		Phone: savedOrder.BankDetail.Phone,
// 		MaxOrdersSimultaneosly: savedOrder.BankDetail.MaxOrdersSimultaneosly,
// 		MaxAmountDay: float64(savedOrder.BankDetail.MaxAmountDay),
// 		MaxAmountMonth: float64(savedOrder.BankDetail.MaxAmountMonth),
// 		MaxQuantityDay: float64(savedOrder.BankDetail.MaxQuantityDay),
// 		MaxQuantityMonth: float64(savedOrder.BankDetail.MaxQuantityMonth),
// 		DeviceId: savedOrder.BankDetail.DeviceID,
// 		InflowCurrency: savedOrder.BankDetail.InflowCurrency,
// 		BankCode: savedOrder.BankDetail.BankCode,
// 		NspkCode: savedOrder.BankDetail.NspkCode,
// 	},
// 	AmountFiat: float64(savedOrder.AmountFiat),
// 	AmountCrypto: float64(savedOrder.AmountCrypto),
// 	ExpiresAt: timestamppb.New(savedOrder.ExpiresAt),
// 	Shuffle: savedOrder.Shuffle,
// 	MerchantOrderId: savedOrder.MerchantOrderID,
// 	ClientId: savedOrder.ClientID,
// 	CallbackUrl: savedOrder.CallbackURL,
// 	TraderRewardPercent: savedOrder.TraderRewardPercent,
// 	CreatedAt: timestamppb.New(savedOrder.CreatedAt),
// 	UpdatedAt: timestamppb.New(savedOrder.UpdatedAt),
// 	Recalculated: savedOrder.Recalculated,
// 	CryptoRubRate: savedOrder.CryptoRubRate,