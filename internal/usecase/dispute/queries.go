package usecase

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	disputedto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/dispute"
)

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

func (disputeUc *DefaultDisputeUsecase) GetDisputeByID(disputeID string) (*domain.Dispute, error) {
	return disputeUc.disputeRepo.GetDisputeByID(disputeID)
}

func (disputeUc *DefaultDisputeUsecase) GetDisputeByOrderID(orderID string) (*domain.Dispute, error) {
	return disputeUc.disputeRepo.GetDisputeByOrderID(orderID)
}