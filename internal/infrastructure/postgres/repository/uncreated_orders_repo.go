package repository

import (
	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/mappers"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
	"gorm.io/gorm"
)

type DefaultUncreatedOrderRepository struct {
	DB *gorm.DB
}

func NewDefaultUncreatedOrderRepository(db *gorm.DB) *DefaultUncreatedOrderRepository {
	return &DefaultUncreatedOrderRepository{
		DB: db,
	}
}

func (r *DefaultUncreatedOrderRepository) CreateLog(log *domain.UncreatedOrder) error {
	model := mappers.ToGORMUncreatedOrder(log)
	return r.DB.Create(model).Error
}

func (r *DefaultUncreatedOrderRepository) GetLogsWithFilters(filter *domain.UncreatedOrdersFilter, page, limit int32, sortBy, sortOrder string) ([]*domain.UncreatedOrder, int64, error) {
	var uncreatedOrderModels []*models.UncreatedOrderModel
	var total int64

	query := r.DB.Model(&models.UncreatedOrderModel{})
	if filter != nil {
		if filter.MerchantID != nil {
			query = query.Where("merchant_id = ?", *filter.MerchantID)
		}
		if filter.MinAmountFiat != nil {
			query = query.Where("amount_fiat >= ?", *filter.MinAmountFiat)
		}
		if filter.MaxAmountFiat != nil {
			query = query.Where("amount_fiat <= ?", *filter.MaxAmountFiat)
		}
		if filter.Currency != nil {
			query = query.Where("currency = ?", *filter.Currency)
		}
		if filter.ClientID != nil {
			query = query.Where("client_id = ?", *filter.ClientID)
		}
		if filter.TimeOpeningStart != nil {
			query = query.Where("created_at >= ?", *filter.TimeOpeningStart)
		}
		if filter.TimeOpeningEnd != nil {
			query = query.Where("created_at <= ?", *filter.TimeOpeningEnd)
		}
		if filter.PaymentSystem != nil {
			query = query.Where("payment_system = ?", *filter.PaymentSystem)
		}
		if filter.BankCode != nil {
			query = query.Where("bank_code = ?", *filter.BankCode)
		}
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	orderClause := sortBy + " " + sortOrder

	if err := query.Order(orderClause).Offset(int(offset)).Limit(int(limit)).Find(&uncreatedOrderModels).Error; err != nil {
		return nil, 0, err
	}

	uncreatedOrders := make([]*domain.UncreatedOrder, len(uncreatedOrderModels))
	for i, uncreatedOrderModel := range uncreatedOrderModels {
		uncreatedOrders[i] = mappers.ToDomainUncreatedOrder(uncreatedOrderModel)
	}

	return uncreatedOrders, total, nil
}
