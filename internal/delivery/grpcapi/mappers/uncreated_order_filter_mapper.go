package mappers

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	orderpb "github.com/LavaJover/shvark-order-service/proto/gen"
)

func ToDomainUncreatedOrdersFilter(filter *orderpb.UncreatedOrderFilters) *domain.UncreatedOrdersFilter {
	var domainFilter *domain.UncreatedOrdersFilter = nil
	if filter != nil {
		domainFilter = &domain.UncreatedOrdersFilter{
			MerchantID:    filter.MerchantId,
			MinAmountFiat: filter.MinAmountFiat,
			MaxAmountFiat: filter.MaxAmountFiat,
			Currency:      filter.Currency,
			ClientID:      filter.ClientId,
			PaymentSystem: filter.PaymentSystem,
			BankCode:      filter.BankCode,
		}

		if filter.DateFrom != nil {
			dateFrom := filter.DateFrom.AsTime()
			domainFilter.TimeOpeningStart = &dateFrom
		}

		if filter.DateTo != nil {
			dateTo := filter.DateTo.AsTime()
			domainFilter.TimeOpeningEnd = &dateTo
		}
	}
	return domainFilter
}
