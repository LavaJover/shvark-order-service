package usecase

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	uncreatedorderdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/uncreatedOrder"
)

type UncreatedOrderUsecase interface {
	LogEvent(event *domain.UncreatedOrder) error
	GetUncreatedLogsWithFilter(filter *domain.UncreatedOrdersFilter, page, limit int32, sortBy, sortOrder string) (*uncreatedorderdto.GetUncreatedOrdersOutput, error)
	GetStatsForUncreatedOrdersWithFilter(filter *domain.UncreatedOrdersFilter, groupByCriteria []string) ([]*domain.UncreatedOrdersStats, error)
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

func (uc *DefaultUncreatedOrderUsecase) GetUncreatedLogsWithFilter(filter *domain.UncreatedOrdersFilter, page, limit int32, sortBy, sortOrder string) (*uncreatedorderdto.GetUncreatedOrdersOutput, error) {
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

func (uc *DefaultUncreatedOrderUsecase) GetStatsForUncreatedOrdersWithFilter(filter *domain.UncreatedOrdersFilter, groupByCriteria []string) ([]*domain.UncreatedOrdersStats, error) {
	safeGroupBy := map[string]bool{
		"merchant_id":        true,
		"currency":           true,
		"payment_system":     true,
		"amount_range_100":   true,
		"amount_range_1000":  true,
		"amount_range_10000": true,
		"date_range_hour":    true,
		"date_range_day":     true,
		"date_range_week":    true,
		"date_range_month":   true,
		"bank_code":          true,
	}

	groupByCriteriaInput := []string{}
	for _, criteria := range groupByCriteria {
		if safeGroupBy[criteria] {
			groupByCriteriaInput = append(groupByCriteriaInput, criteria)
		}
	}

	output, err := uc.uncreatedOrdersRepo.GetStatsWithFIlters(filter, groupByCriteriaInput)

	if err != nil {
		return nil, err
	}

	return output, nil
}
