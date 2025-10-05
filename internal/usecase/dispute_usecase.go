package usecase

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	walletRequest "github.com/LavaJover/shvark-order-service/internal/delivery/http/dto/wallet/request"
	"github.com/LavaJover/shvark-order-service/internal/delivery/http/handlers"
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/bitwire/notifier"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/kafka"
	disputedto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/dispute"
	"github.com/jaevor/go-nanoid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DisputeUsecase interface {
	CreateDispute(input *disputedto.CreateDisputeInput) error
	AcceptDispute(disputeID string) error
	RejectDispute(disputeID string) error
	FreezeDispute(disputeID string) error
	GetDisputeByID(disputeID string) (*domain.Dispute, error)
	GetDisputeByOrderID(orderID string) (*domain.Dispute, error)
	AcceptExpiredDisputes() error
	GetOrderDisputes(input *disputedto.GetOrderDisputesInput) (*disputedto.GetOrderDisputesOutput, error)
}

type DefaultDisputeUsecase struct {
	disputeRepo domain.DisputeRepository
	walletHandler *handlers.HTTPWalletHandler
	orderRepo domain.OrderRepository
	trafficRepo domain.TrafficRepository
	kafkaPublisher *publisher.KafkaPublisher
	teamRelationsUsecase TeamRelationsUsecase
	bankDetailRepo domain.BankDetailRepository
}

func NewDefaultDisputeUsecase(
	disputeRepo domain.DisputeRepository,
	walletHandler *handlers.HTTPWalletHandler,
	orderRepo domain.OrderRepository,
	trafficRepo domain.TrafficRepository,
	kafkaPublisher *publisher.KafkaPublisher,
	teamRelationsUsecase TeamRelationsUsecase,
	bankDetailRepo domain.BankDetailRepository,
	) *DefaultDisputeUsecase {
	return &DefaultDisputeUsecase{
		disputeRepo: disputeRepo,
		walletHandler: walletHandler,
		orderRepo: orderRepo,
		trafficRepo: trafficRepo,
		kafkaPublisher: kafkaPublisher,
		teamRelationsUsecase: teamRelationsUsecase,
		bankDetailRepo: bankDetailRepo,
	}
}

// Диспут открыт -> запись в БД со статусом DISPUTE_OPENED
// AutoAcceptAt -> в данное время система автоматически одобрит диспут
func (disputeUc *DefaultDisputeUsecase) CreateDispute(input *disputedto.CreateDisputeInput) error {
	order, err := disputeUc.orderRepo.GetOrderByID(input.OrderID)
	if err != nil {
		return err
	}
	if (order.Status != domain.StatusCanceled) && (order.Status != domain.StatusCompleted) {
		return status.Error(codes.FailedPrecondition, "invalid order status")
	}
	idGenerator, err := nanoid.Standard(15)
	if err != nil {
		return err
	}
	dispute := domain.Dispute{
		ID: idGenerator(),
		OrderID: input.OrderID,
		OrderStatusOriginal: order.Status,
		DisputeAmountFiat: input.DisputeAmountFiat,
		DisputeAmountCrypto: input.DisputeAmountCrypto,
		DisputeCryptoRate: input.DisputeCryptoRate,
		ProofUrl: input.ProofUrl,
		Reason: input.Reason,
		Status: domain.DisputeOpened,
		Ttl: input.Ttl,
		AutoAcceptAt: time.Now().Add(input.Ttl),
	}

	bankDetailID := order.BankDetailID
	bankDetail, err := disputeUc.bankDetailRepo.GetBankDetailByID(bankDetailID)
	if err != nil {
		return err
	}
	err = disputeUc.disputeRepo.CreateDispute(&dispute)
	if err != nil {
		return err
	}
	go func(event publisher.DisputeEvent){
		if err := disputeUc.kafkaPublisher.PublishDispute(event); err != nil {
			slog.Error("failed to publish kafka dispute event", "stage", "creating", "error",err.Error())
		}
	}(publisher.DisputeEvent{
		DisputeID: dispute.ID,
		OrderID: dispute.OrderID,
		TraderID: bankDetail.TraderID,
		OrderAmountFiat: order.AmountInfo.AmountFiat,
		DisputeAmountFiat: dispute.DisputeAmountFiat,
		ProofUrl: dispute.ProofUrl,
		Reason: dispute.Reason,
		Status: "🆘Открыт диспут",
		BankName: bankDetail.BankName,
		Phone: bankDetail.Phone,
		CardNumber: bankDetail.CardNumber,
		Owner: bankDetail.Owner,
	})

	if order.Status == domain.StatusCompleted {
		err = disputeUc.walletHandler.Freeze(bankDetail.TraderID, fmt.Sprintf("%s_dispute_%s", dispute.OrderID, dispute.ID), dispute.DisputeAmountCrypto-order.AmountInfo.AmountCrypto)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}else {
		err = disputeUc.walletHandler.Freeze(bankDetail.TraderID, fmt.Sprintf("%s_dispute_%s", dispute.OrderID, dispute.ID), dispute.DisputeAmountCrypto)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}
	err = disputeUc.orderRepo.UpdateOrderStatus(order.ID, domain.StatusDisputeCreated)
	if err != nil {
		return err
	}
	if order.CallbackUrl != "" {
		notifier.SendCallback(
			order.CallbackUrl,
			order.MerchantInfo.MerchantOrderID,
			string(domain.StatusDisputeCreated),
			0, 0, 0,
		)
	}
	return nil
}

func (disputeUc *DefaultDisputeUsecase) AcceptDispute(disputeID string) error {
	dispute, err := disputeUc.disputeRepo.GetDisputeByID(disputeID)
	if err != nil {
		return err
	}
	order, err := disputeUc.orderRepo.GetOrderByID(dispute.OrderID)
	if err != nil {
		return err
	}
	if order.Status != domain.StatusDisputeCreated {
		return fmt.Errorf("invalid order status to accept dispute: %s", order.Status)
	}
	bankDetailID := order.BankDetailID
	bankDetail, err := disputeUc.bankDetailRepo.GetBankDetailByID(bankDetailID)
	if err != nil {
		return err
	}
	traffic, err := disputeUc.trafficRepo.GetTrafficByTraderMerchant(bankDetail.TraderID, order.MerchantInfo.MerchantID)
	if err != nil {
		return err
	}
	// Search for team relations to find commission users
	var commissionUsers []walletRequest.CommissionUser
	teamRelations, err := disputeUc.teamRelationsUsecase.GetRelationshipsByTraderID(bankDetail.TraderID)
	if err == nil {
		for _, teamRelation := range teamRelations {
			commissionUsers = append(commissionUsers, walletRequest.CommissionUser{
				UserID: teamRelation.TeamLeadID,
				Commission: teamRelation.TeamRelationshipRapams.Commission,
			})
		}
	}
	op := &DisputeOperation{
		OrderID: order.ID,
		DisputeID: disputeID,
		Operation: "accept",
		OldDisputeStatus: dispute.Status,
		NewDisputeStatus: domain.DisputeAccepted,
		OldOrderStatus: order.Status,
		NewOrderStatus: domain.StatusCompleted,
		NewOrderAmountFiat: dispute.DisputeAmountFiat,
		NewOrderAmountCrypto: dispute.DisputeAmountCrypto,
		NewOrderAmountCryptoRate: dispute.DisputeCryptoRate,		
		WalletOp: &WalletOperation{
			Type: "release",
			Request: walletRequest.ReleaseRequest{
				TraderID: bankDetail.TraderID,
				MerchantID: order.MerchantInfo.MerchantID,
				OrderID: fmt.Sprintf("%s_dispute_%s", dispute.OrderID, dispute.ID),
				RewardPercent: traffic.TraderRewardPercent,
				PlatformFee: traffic.PlatformFee,
				CommissionUsers: commissionUsers,
			},
		},
		CreatedAt: time.Now(),
	}

	if err := disputeUc.ProcessDisputeOperation(context.Background(), op); err != nil {
		return err
	}
	
	if order.CallbackUrl != "" {
		notifier.SendCallback(
			order.CallbackUrl,
			order.MerchantInfo.MerchantOrderID,
			string(domain.StatusCompleted),
			dispute.DisputeAmountCrypto, dispute.DisputeAmountFiat, dispute.DisputeCryptoRate,
		)
	}
	return nil
}

func (disputeUc *DefaultDisputeUsecase) RejectDispute(disputeID string) error {
	dispute, err := disputeUc.disputeRepo.GetDisputeByID(disputeID)
	if err != nil {
		return err
	}
	if dispute.Status != domain.DisputeOpened && dispute.Status != domain.DisputeFreezed {
		return fmt.Errorf("invalid dispute status to reject dispute: %s", dispute.Status)
	}
	order, err := disputeUc.orderRepo.GetOrderByID(dispute.OrderID)
	if err != nil {
		return err
	}
	if order.Status != domain.StatusDisputeCreated {
		return fmt.Errorf("invalid order status to reject dispute: %s", order.Status)
	}
	bankDetailID := order.BankDetailID
	bankDetail, err := disputeUc.bankDetailRepo.GetBankDetailByID(bankDetailID)
	if err != nil {
		return err
	}
	op := &DisputeOperation{
		OrderID: order.ID,
		DisputeID: disputeID,
		Operation: "reject",
		OldDisputeStatus: dispute.Status,
		NewDisputeStatus: domain.DisputeRejected,
		OldOrderStatus: order.Status,
		NewOrderStatus: dispute.OrderStatusOriginal,
		WalletOp: &WalletOperation{
			Type: "release",
			Request: walletRequest.ReleaseRequest{
				TraderID: bankDetail.TraderID,
				MerchantID: order.MerchantInfo.MerchantID,
				OrderID: fmt.Sprintf("%s_dispute_%s", dispute.OrderID, dispute.ID),
				RewardPercent: 1,
				PlatformFee: 1,
			},
		},
		CreatedAt: time.Now(),
	}
	if err := disputeUc.ProcessDisputeOperation(context.Background(), op); err != nil {
		return err
	}
	if order.CallbackUrl != "" {
		notifier.SendCallback(
			order.CallbackUrl,
			order.MerchantInfo.MerchantOrderID,
			string(domain.StatusCanceled),
			0, 0, 0,
		)
	}
	
	return nil
}

func (disputeUc *DefaultDisputeUsecase) FreezeDispute(disputeID string) error {
	dispute, err := disputeUc.disputeRepo.GetDisputeByID(disputeID)
	if err != nil {
		return err
	}
	if dispute.Status != domain.DisputeOpened {
		return status.Error(codes.FailedPrecondition, "dispute is not opened yet")
	}

	op := &DisputeOperation{
		OrderID: dispute.OrderID,
		OldOrderStatus: domain.StatusDisputeCreated,
		NewOrderStatus: domain.StatusDisputeCreated,
		DisputeID: dispute.ID,
		Operation: "freeze",
		OldDisputeStatus: dispute.Status,
		NewDisputeStatus: domain.DisputeFreezed,
		WalletOp: nil,
		CreatedAt: time.Now(),
	}

	return disputeUc.ProcessDisputeOperation(context.Background(), op)
}

func (disputeUc *DefaultDisputeUsecase) GetDisputeByID(disputeID string) (*domain.Dispute, error) {
	return disputeUc.disputeRepo.GetDisputeByID(disputeID)
}

func (disputeUc *DefaultDisputeUsecase) GetDisputeByOrderID(orderID string) (*domain.Dispute, error) {
	return disputeUc.disputeRepo.GetDisputeByOrderID(orderID)
}

func (disputeUc *DefaultDisputeUsecase) AcceptExpiredDisputes() error {
	disputes, err := disputeUc.disputeRepo.FindExpiredDisputes()
	if err != nil {
		return err
	}
	for _, dispute := range disputes {
		if err := disputeUc.AcceptDispute(dispute.ID); err != nil {
			log.Printf("failed to accept dispute %s\n", dispute.ID)
			return status.Error(codes.Internal, err.Error())
		}
	} 
	return nil
}

func (disputeUc *DefaultDisputeUsecase) GetOrderDisputes(input *disputedto.GetOrderDisputesInput) (*disputedto.GetOrderDisputesOutput, error) {
	filter := domain.GetDisputesFilter{
		DisputeID: input.DisputeID,
		TraderID: input.TraderID,
		OrderID: input.OrderID,
		MerchantID: input.MerchantID,
		Status: input.Status,
		Page: int(input.Page),
		Limit: int(input.Limit),
	}
	disputes, total, err := disputeUc.disputeRepo.GetOrderDisputes(filter)
	if err != nil {
		return nil, err
	}

	totalPages := total / input.Limit
	if total % input.Limit != 0 {
		totalPages++
	}

	return &disputedto.GetOrderDisputesOutput{
		Disputes: disputes,
		Pagination: disputedto.Pagination{
			CurrentPage: int32(input.Page),
			TotalPages: int32(totalPages),
			TotalItems: int32(total),
			ItemsPerPage: int32(input.Limit),
		},
	}, nil
}

////////////////////// Advanced Safe Dispute operations //////////////////////////

// DisputeOperation - описание операции с диспутом
type DisputeOperation struct {
	OrderID		string								`json:"order_id"`
    DisputeID   string                   			`json:"dispute_id"`
    Operation   string                    			`json:"operation"` // "create", "approve", "cancel", "freeze"
	OldOrderStatus domain.OrderStatus				`json:"old_order_status"`
	NewOrderStatus domain.OrderStatus				`json:"new_order_status"`
    OldDisputeStatus   domain.DisputeStatus        	`json:"old_status"`
    NewDisputeStatus   domain.DisputeStatus        	`json:"new_status"`
	NewOrderAmountFiat float64
	NewOrderAmountCrypto float64
	NewOrderAmountCryptoRate float64
    WalletOp    *WalletOperation         			`json:"wallet_op,omitempty"`
    CreatedAt   time.Time                			`json:"created_at"`
}

///////////////////////// Базовая транзакционная функция //////////////////////////

// ProcessOrderOperation - базовая функция для всех операций со сделками
func (disputeUc *DefaultDisputeUsecase) ProcessDisputeOperation(ctx context.Context, op *DisputeOperation) error {
    // 1. КРИТИЧНО: Атомарно меняем статус и обрабатываем кошелек
    if err := disputeUc.processCriticalOperations(ctx, op); err != nil {
        return fmt.Errorf("critical operations failed: %w", err)
    }

    // 2. НЕКРИТИЧНО: Асинхронно публикуем событие и отправляем callback
    // if err := uc.scheduleNonCriticalOperations(op); err != nil {
    //     log.Printf("Failed to schedule non-critical operations for order %s: %v", op.OrderID, err)
    //     // НЕ возвращаем ошибку - критичные операции уже выполнены
    // }

    return nil
}

// processCriticalOperations - синхронная обработка критичных операций
func (disputeUc *DefaultDisputeUsecase) processCriticalOperations(ctx context.Context, op *DisputeOperation) error {
    var walletFunc func() error
    if op.WalletOp != nil {
        walletFunc = func() error {
            return disputeUc.processWalletOperation(op.WalletOp)
        }
    }

    return disputeUc.disputeRepo.ProcessDisputeCriticalOperation(
        op.DisputeID, 
		op.OrderID,
        op.NewDisputeStatus,
		op.NewOrderStatus,
		op.NewOrderAmountFiat, op.NewOrderAmountCrypto, op.NewOrderAmountCryptoRate,
        op.Operation, // передаем тип операции
        walletFunc,
    )
}

// processWalletOperation - обработка операций с кошельком
func (disputeUc *DefaultDisputeUsecase) processWalletOperation(walletOp *WalletOperation) error {
    switch walletOp.Type {
    case "freeze":
        req := walletOp.Request.(walletRequest.FreezeRequest)
        return disputeUc.walletHandler.Freeze(req.TraderID, req.OrderID, req.Amount)
    case "release":
        req := walletOp.Request.(walletRequest.ReleaseRequest)
        return disputeUc.walletHandler.Release(req)
    default:
        return fmt.Errorf("unknown wallet operation: %s", walletOp.Type)
    }
}