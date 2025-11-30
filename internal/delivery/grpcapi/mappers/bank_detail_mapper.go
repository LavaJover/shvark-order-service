package mappers

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen/order"
	"google.golang.org/protobuf/types/known/durationpb"
)

func ToProtoBankDetail(bankDetail *domain.BankDetail) *orderpb.BankDetail {
	return &orderpb.BankDetail{
		BankDetailId: bankDetail.ID,
		TraderId: bankDetail.TraderID,
		Currency: bankDetail.Currency,
		Country: bankDetail.Country,
		MinAmount: float64(bankDetail.MinOrderAmount),
		MaxAmount: float64(bankDetail.MaxOrderAmount),
		BankName: bankDetail.BankName,
		PaymentSystem: bankDetail.PaymentSystem,
		Enabled: bankDetail.Enabled,
		Delay: durationpb.New(bankDetail.Delay),
		CardNumber: bankDetail.CardNumber,
		Phone: bankDetail.Phone,
		Owner: bankDetail.Owner,
		MaxOrdersSimultaneosly: bankDetail.MaxOrdersSimultaneosly,
		MaxAmountDay: bankDetail.MaxAmountDay,
		MaxAmountMonth: bankDetail.MaxAmountMonth,
		MaxQuantityDay: float64(bankDetail.MaxQuantityDay),
		MaxQuantityMonth: float64(bankDetail.MaxQuantityMonth),
		DeviceId: bankDetail.DeviceID,
		InflowCurrency: bankDetail.InflowCurrency,
		BankCode: bankDetail.BankCode,
		NspkCode: bankDetail.NspkCode,
	}
}