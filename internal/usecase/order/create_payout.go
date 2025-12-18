package usecase

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/bitwire/notifier"
	publisher "github.com/LavaJover/shvark-order-service/internal/infrastructure/kafka"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/usdt"
	orderdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/order"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (uc *DefaultOrderUsecase) pickTraderForPayOut(trafficRecords []*domain.Traffic) (*domain.Traffic, error) {
    // Фильтруем активных трейдеров с проверкой всех условий
    activeTraders := make([]*domain.Traffic, 0, len(trafficRecords))
    var totalPriority float64

    for _, traffic := range trafficRecords {
        if !traffic.Enabled {
            continue
        }

        activity := traffic.ActivityParams
        if !activity.MerchantUnlocked || !activity.TraderUnlocked || 
           !activity.AntifraudUnlocked || !activity.ManuallyUnlocked {
            continue
        }

        if traffic.TraderPriority <= 0 {
            continue
        }

        activeTraders = append(activeTraders, traffic)
        totalPriority += traffic.TraderPriority
    }

    if len(activeTraders) == 0 {
        return nil, fmt.Errorf("no available traders")
    }

    if len(activeTraders) == 1 {
        return activeTraders[0], nil
    }

    // Взвешенный случайный выбор
    randomValue := rand.Float64() * totalPriority
    var cumulativePriority float64

    for _, traffic := range activeTraders {
        cumulativePriority += traffic.TraderPriority
        if randomValue <= cumulativePriority {
            return traffic, nil
        }
    }

    // Фоллбэк на последнего трейдера (достижимо только при ошибках округления)
    return activeTraders[len(activeTraders)-1], nil
}

func (uc *DefaultOrderUsecase) CreatePayOutOrder (createOrderInput *orderdto.CreatePayOutOrderInput) (*orderdto.OrderOutput, error) {
    slog.Info("CreatePayOutOrder started")
    
	// Проверить баланс мерча
	merchantID := createOrderInput.MerchantID
	cryptoRate := usdt.UsdtRubRates
	amountCrypto := createOrderInput.AmoutFiat / usdt.UsdtRubRates
	balance, err := uc.WalletHandler.GetTraderBalance(merchantID)
	if err != nil {
		slog.Error("failed to fetch merchant account balance", "error", err.Error())
		return nil, err
	}
	if balance < amountCrypto {
		slog.Error("failed to create pay out order", "msg", "merchant account balance too low", "balance", balance, "order_amount_crupto", amountCrypto)
		return nil, status.Errorf(codes.FailedPrecondition, "Low balance: %f. Needed atleast: %f", balance, amountCrypto)
	}

	// Выбрать трейдера
	trafficRecords, err := uc.TrafficUsecase.GetTrafficByMerchantID(merchantID)
	if err != nil {
		slog.Error("failed to get traffic records", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "Failed to fetch traffic records")
	}

	chosenTraffic, err := uc.pickTraderForPayOut(trafficRecords)
	if err != nil {
		slog.Error("failed to pick trader for pay out", "error", err.Error())
		return nil, status.Errorf(codes.NotFound, "no available traders active")
	}

    if createOrderInput.AdvancedParams.CallbackUrl != "" {
        notifier.SendCallback(
            createOrderInput.AdvancedParams.CallbackUrl,
            createOrderInput.MerchantOrderID,
            string(domain.StatusCreated),
            0, 0, 0,
        )
    }

    traderReward := chosenTraffic.TraderRewardPercent
    platformFee := chosenTraffic.PlatformFee

    order := domain.Order{
        ID:     uuid.New().String(),
        Status: domain.StatusCreated,
        MerchantInfo: domain.MerchantInfo{
            MerchantID:     createOrderInput.MerchantID,
            MerchantOrderID: createOrderInput.MerchantOrderID,
            ClientID:       createOrderInput.ClientID,
        },
        AmountInfo: domain.AmountInfo{
            AmountFiat:   createOrderInput.AmoutFiat,
            AmountCrypto: amountCrypto,
            CryptoRate:   cryptoRate,
            Currency:     createOrderInput.Currency,
        },
        BankDetailID:  nil,
        Type:          domain.TypePayOut,
        Recalculated:  createOrderInput.Recalculated,
        Shuffle:       createOrderInput.Shuffle,
        TraderReward:  traderReward,
        PlatformFee:   platformFee,
        CallbackUrl:   createOrderInput.CallbackUrl,
        ExpiresAt:     time.Now().Add(chosenTraffic.BusinessParams.MerchantDealsDuration),

        RequisiteDetails: domain.RequisiteDetails{
            TraderID: chosenTraffic.TraderID,
            CardNumber: createOrderInput.PaymentDetails.CardNumber,
            Phone: createOrderInput.PaymentDetails.Phone,
            Owner: createOrderInput.PaymentDetails.Owner,
            PaymentSystem: createOrderInput.PaymentDetails.PaymentSystem,
            BankName: createOrderInput.PaymentDetails.BankInfo.BankName,
            BankCode: createOrderInput.PaymentDetails.BankInfo.BankCode,
            NspkCode: createOrderInput.PaymentDetails.BankInfo.NspkCode,
            DeviceID: "",
        },
        Metrics: domain.Metrics{},
    }
    
    err = uc.OrderRepo.CreateOrder(&order)
    if err != nil {
        return nil, err
    }

    // Freeze crypto
	// Замораживаем у мерчанта
    if err := uc.WalletHandler.Freeze(merchantID, order.ID, amountCrypto); err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    // Publish to kafka асинхронно
    go func(event publisher.OrderEvent) {
        if err := uc.Publisher.PublishOrder(event); err != nil {
            slog.Error("failed to publish OrderEvent:created", "error", err.Error())
        }
    }(publisher.OrderEvent{
        OrderID:   order.ID,
        TraderID:  order.RequisiteDetails.TraderID,
        Status:    "🔥Новая Выплата",
        AmountFiat: order.AmountInfo.AmountFiat,
        Currency:  order.AmountInfo.Currency,
        BankName:  order.RequisiteDetails.BankName,
        Phone:     order.RequisiteDetails.Phone,
        CardNumber: order.RequisiteDetails.CardNumber,
        Owner:     order.RequisiteDetails.Owner,
    })

    return &orderdto.OrderOutput{
        Order:     order,
        BankDetail: domain.BankDetail{
			ID: "",
			SearchParams: domain.SearchParams{},
			DeviceInfo: domain.DeviceInfo{},
			TraderInfo: domain.TraderInfo{
				TraderID: chosenTraffic.TraderID,
			},
			PaymentDetails: domain.PaymentDetails{
				Phone: order.RequisiteDetails.Phone,
				CardNumber: order.RequisiteDetails.CardNumber,
				Owner: order.RequisiteDetails.Owner,
				PaymentSystem: order.RequisiteDetails.PaymentSystem,
				BankInfo: domain.BankInfo{
					BankCode: order.RequisiteDetails.BankCode,
					BankName: order.RequisiteDetails.BankName,
					NspkCode: order.RequisiteDetails.NspkCode,
				},
			},
			Country: "",
			Currency: order.AmountInfo.Currency,
			InflowCurrency: "",
			CreatedAt: order.CreatedAt,
			UpdatedAt: order.UpdatedAt,
		},
    }, nil
}