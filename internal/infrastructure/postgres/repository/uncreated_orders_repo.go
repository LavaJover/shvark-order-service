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
	applyFilters(query, filter)

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

func (r *DefaultUncreatedOrderRepository) GetStatsWithFIlters(filter *domain.UncreatedOrdersFilter, groupByCriteria []string) ([]*domain.UncreatedOrdersStats, error) {
	query := r.DB.Model(&models.UncreatedOrderModel{})
	applyFilters(query, filter)

	selectFields := buildSelectFields(groupByCriteria)
	query = query.Select(selectFields)

	if len(groupByCriteria) > 0 {
		groupByFields := buildGroupByFields(groupByCriteria)
		query.Group(groupByFields)
	}

	var stats []domain.UncreatedOrdersStats
	if err := query.Find(&stats).Error; err != nil {
		return nil, err
	}

	result := make([]*domain.UncreatedOrdersStats, len(stats))
	for i := range stats {
		result[i] = &stats[i]
	}
	return result, nil
}

func applyFilters(query *gorm.DB, filter *domain.UncreatedOrdersFilter) {
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
}

func buildSelectFields(groupByCriteria []string) string {
	baseAggregations := `
		COUNT(*) as total_count,
		COALESCE(SUM(amount_fiat), 0) as total_amount_fiat,
		COALESCE(AVG(amount_fiat), 0) as avg_amount_fiat,
		COALESCE(MIN(amount_fiat), 0) as min_amount_fiat,
		COALESCE(MAX(amount_fiat), 0) as max_amount_fiat`

	if len(groupByCriteria) == 0 {
		return baseAggregations
	}

	selectFields := ""
	for _, criteria := range groupByCriteria {
		switch criteria {
		case "merchant_id":
			selectFields += "merchant_id, "
		case "currency":
			selectFields += "currency, "
		case "payment_system":
			selectFields += "payment_system, "
		case "bank_code":
			selectFields += "bank_code, "
		case "amount_range_100":
			selectFields += "FLOOR(amount_fiat / 100) * 100 as amount_range, "
		case "amount_range_1000":
			selectFields += "FLOOR(amount_fiat / 1000) * 1000 as amount_range, "
		case "amount_range_10000":
			selectFields += "FLOOR(amount_fiat / 10000) * 10000 as amount_range, "
		case "date_range_hour":
			selectFields += "DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00') as date_group, "
		case "date_range_day":
			selectFields += "DATE(created_at) as date_group, "
		case "date_range_week":
			selectFields += "DATE_SUB(DATE(created_at), INTERVAL WEEKDAY(created_at) DAY) as date_group, "
		case "date_range_month":
			selectFields += "DATE_FORMAT(created_at, '%Y-%m-01') as date_group, "
		}
	}
	return selectFields + baseAggregations
}

func buildGroupByFields(groupByCriteria []string) string {
	groupByFields := ""

	for i, criteria := range groupByCriteria {
		if i > 0 {
			groupByFields += ", "
		}
		switch criteria {
		case "merchant_id":
			groupByFields += "merchant_id"
		case "currency":
			groupByFields += "currency"
		case "payment_system":
			groupByFields += "payment_system"
		case "bank_code":
			groupByFields += "bank_code"
		case "date_range_hour":
			groupByFields += "DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00')"
		case "date_range_day":
			groupByFields += "DATE(created_at)"
		case "date_range_week":
			groupByFields += "DATE_SUB(DATE(created_at), INTERVAL WEEKDAY(created_at) DAY)"
		case "date_range_month":
			groupByFields += "DATE_FORMAT(created_at, '%Y-%m-01')"
		case "amount_range_100":
			groupByFields += "FLOOR(amount_fiat / 100) * 100"
		case "amount_range_1000":
			groupByFields += "FLOOR(amount_fiat / 1000) * 1000"
		case "amount_range_10000":
			groupByFields += "FLOOR(amount_fiat / 10000) * 10000"
		}
	}

	return groupByFields
}
