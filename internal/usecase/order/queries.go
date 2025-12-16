package usecase

import (
	"fmt"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	orderdto "github.com/LavaJover/shvark-order-service/internal/usecase/dto/order"
)

func (uc *DefaultOrderUsecase) GetOrderStatistics(traderID string, dateFrom, dateTo time.Time) (*domain.OrderStatistics, error) {
	return uc.OrderRepo.GetOrderStatistics(traderID, dateFrom, dateTo)
}

func (uc *DefaultOrderUsecase) GetOrders(filter domain.Filter, sortField string, page, size int) ([]*domain.Order, int64, error) {
	return uc.OrderRepo.GetOrders(filter, sortField, page, size)
}

func (uc *DefaultOrderUsecase) GetAllOrders(input *orderdto.GetAllOrdersInput) (*orderdto.GetAllOrdersOutput, error) {
    // Валидация пагинации
    if input.Page < 1 {
        input.Page = 1
    }
    if input.Limit < 1 || input.Limit > 100 {
        input.Limit = 50 // дефолтное значение
    }

    // Преобразуем в фильтры репозитория
    filters := &domain.AllOrdersFilters{
        TraderID:         input.TraderID,
        MerchantID:       input.MerchantID,
        OrderID:          input.OrderID,
        MerchantOrderID:  input.MerchantOrderID,
        Status:           input.Status,
        BankCode:         input.BankCode,
        TimeOpeningStart: input.TimeOpeningStart,
        TimeOpeningEnd:   input.TimeOpeningEnd,
        AmountFiatMin:    input.AmountFiatMin,
        AmountFiatMax:    input.AmountFiatMax,
        Type:             input.Type,
        DeviceID:         input.DeviceID,
		PaymentSystem: 	  input.PaymentSystem,
    }

    // Вызываем репозиторий
    orders, total, err := uc.OrderRepo.GetAllOrders(filters, input.Sort, input.Page, input.Limit)
    if err != nil {
        return nil, err
    }

    // Рассчитываем данные пагинации
    totalPages := int32(total) / input.Limit
    if int32(total)%input.Limit > 0 {
        totalPages++
    }

    return &orderdto.GetAllOrdersOutput{
        Orders: orders,
        Pagination: orderdto.Pagination{
            CurrentPage:  input.Page,
            TotalPages:   totalPages,
            TotalItems:   int32(total),
            ItemsPerPage: input.Limit,
        },
    }, nil
}

func (uc *DefaultOrderUsecase) GetOrderByID(orderID string) (*domain.Order, error) {
	return uc.OrderRepo.GetOrderByID(orderID)
}

func (uc *DefaultOrderUsecase) GetOrderByMerchantOrderID(merchantOrderID string) (*domain.Order, error) {
	return uc.OrderRepo.GetOrderByMerchantOrderID(merchantOrderID)
}

func (uc *DefaultOrderUsecase) GetOrdersByTraderID(
	traderID string, page, 
	limit int64, sortBy, 
	sortOrder string,
	filters domain.OrderFilters,
) ([]*orderdto.OrderOutput, int64, error) {

	validStatuses := map[string]bool{
		string(domain.StatusCompleted): true,
		string(domain.StatusCanceled): true,
		string(domain.StatusPending): true,
		string(domain.StatusDisputeCreated): true,
	}

	for _, status := range filters.Statuses {
		if !validStatuses[status] {
			return nil, 0, fmt.Errorf("invalid status in filters")
		}
	}

	orders, total, err := uc.OrderRepo.GetOrdersByTraderID(
		traderID, 
		page, 
		limit, 
		sortBy, 
		sortOrder,
		filters,
	)
	if err != nil {
		return nil, 0, err
	}
	var orderOutputs []*orderdto.OrderOutput
	for _, order := range orders {
		bankDetailID := order.BankDetailID
		bankDetail, err := uc.BankDetailUsecase.GetBankDetailByID(*bankDetailID)
		if err != nil {
			return nil, 0, err
		}
		orderOutputs = append(orderOutputs, &orderdto.OrderOutput{
			Order: *order,
			BankDetail: *bankDetail,
		})
	}

	return orderOutputs, total, nil
}

func (uc *DefaultOrderUsecase) FindExpiredOrders() ([]*domain.Order, error) {
	return uc.OrderRepo.FindExpiredOrders()
}