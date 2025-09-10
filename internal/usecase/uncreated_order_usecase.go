package usecase

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	uncreatedorderdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/uncreatedOrder"
)

type UncreatedOrderUsecase interface {
	LogEvent(event *domain.UncreatedOrder) error
	GetUncreatedLogsWithFilters(filter *domain.UncreatedOrdersFilter, page, limit int32, sortBy, sortOrder string) (*uncreatedorderdto.GetUncreatedOrdersOutput, error)
}

type DefaultUncreatedOrderUsecase struct {
	uncreatedOrdersRepo domain.UncreatedOrderRepository
}

func NewDefaultUncreatedOrderUsecase(uncreatedOrdersRepo domain.UncreatedOrderRepository) *DefaultUncreatedOrderUsecase {
	return &DefaultUncreatedOrderUsecase{
		uncreatedOrdersRepo: uncreatedOrdersRepo,
	}
}

func (uc *DefaultUncreatedOrderUsecase) LogEvent(event *domain.UncreatedOrder) error {
	if err := uc.uncreatedOrdersRepo.CreateLog(event); err != nil {
		return err
	}
	return nil
}

func (uc *DefaultUncreatedOrderUsecase) GetUncreatedLogsWithFilters(filter *domain.UncreatedOrdersFilter, page, limit int32, sortBy, sortOrder string) (*uncreatedorderdto.GetUncreatedOrdersOutput, error) {
	safeSortBy := map[string]bool{
		"merchant_id":   true,
		"amount_fiat":   true,
		"amount_crypto": true,
		"currency":      true,

		"created_at": true,
	}

	if !safeSortBy[sortBy] {
		sortBy = "created_at"
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 1
	}

	data, total, err := uc.uncreatedOrdersRepo.GetLogsWithFilters(filter, page, limit, sortBy, sortOrder)
	if err != nil {
		return nil, err
	}

	totalPages := int32(total) / limit
	if int32(total)%limit > 0 {
		totalPages++
	}

	return &uncreatedorderdto.GetUncreatedOrdersOutput{
		UncreatedOrders: data,
		Pagination: uncreatedorderdto.Pagination{
			CurrentPage:  page,
			TotalPages:   totalPages,
			TotalItems:   int32(total),
			ItemsPerPage: limit,
		},
	}, nil
}
