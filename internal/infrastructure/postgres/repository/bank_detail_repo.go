package repository

import (
	"fmt"
	"time"

	"github.com/LavaJover/shvark-order-service/internal/domain"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/mappers"
	"github.com/LavaJover/shvark-order-service/internal/infrastructure/postgres/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DefaultBankDetailRepo struct {
	DB *gorm.DB
}

func NewDefaultBankDetailRepo(db *gorm.DB) *DefaultBankDetailRepo {
	return &DefaultBankDetailRepo{DB: db}
}

func (r *DefaultBankDetailRepo) SaveBankDetail(bankDetail *domain.BankDetail) error {
	bankDetailModel := mappers.ToGORMBankDetail(bankDetail)
	return r.DB.Save(bankDetailModel).Error
}

func (r *DefaultBankDetailRepo) CreateBankDetail(bankDetail *domain.BankDetail) error {
	bankDetailModel := mappers.ToGORMBankDetail(bankDetail)
	bankDetailModel.ID = uuid.New().String()

	if err := r.DB.Create(bankDetailModel).Error; err != nil {
		return err
	}

	bankDetail.ID = bankDetailModel.ID

	return nil
}

func (r *DefaultBankDetailRepo) UpdateBankDetail(bankDetail *domain.BankDetail) error {
	bankDetailModel := mappers.ToGORMBankDetail(bankDetail)

	if err := r.DB.Model(&models.BankDetailModel{}).Where("id = ?", bankDetailModel.ID).Updates(bankDetailModel).Error; err != nil {
		return err
	}

	if err := r.DB.Model(&models.BankDetailModel{}).Where("id = ?", bankDetailModel.ID).Updates(map[string]interface{}{
		"enabled": bankDetail.Enabled,
		"delay": bankDetail.Delay,
	}).Error; err != nil {
		return err
	}

	return nil
}

func (r *DefaultBankDetailRepo) DeleteBankDetail(bankDetailID string) error {
	return r.DB.Where("id = ?", bankDetailID).Delete(&models.BankDetailModel{}).Error
}

func (r *DefaultBankDetailRepo) GetBankDetailByID(bankDetailID string) (*domain.BankDetail, error) {
	var bankDetailModel models.BankDetailModel
	if err := r.DB.Unscoped().Where("id = ?", bankDetailID).Find(&bankDetailModel).Error; err != nil {
		return nil, err
	}
	return mappers.ToDomainBankDetail(&bankDetailModel), nil
}

func (r *DefaultBankDetailRepo) GetBankDetailsByTraderID(
	traderID string, 
	page, limit int,
	sortOrder, sortBy string,
) ([]*domain.BankDetail, int64, error) {
	var bankDetailModels []*models.BankDetailModel
	var total int64
	
	safeSortBy := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"min_amount": true,
		"max_amount": true,
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

	offset := (page - 1) * limit
	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)

	err := r.DB.
		Model(&models.BankDetailModel{}).
		Where("trader_id = ?", traderID).
		Order(orderClause).
		Limit(limit).
		Offset(offset).
		Find(&bankDetailModels).Error
	
	if err != nil {
		return nil, 0, err
	}

	if err := r.DB.Model(&models.BankDetailModel{}).Where("trader_id = ?", traderID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	bankDetails := make([]*domain.BankDetail, len(bankDetailModels))
	for i, bankDetailModel := range bankDetailModels{
		bankDetails[i] = mappers.ToDomainBankDetail(bankDetailModel)
	}

	return bankDetails, total, nil
}

func (r *DefaultBankDetailRepo) FindSuitableBankDetails(searchQuery *domain.SuitablleBankDetailsQuery) ([]*domain.BankDetail, error) {
    // Этап 1: Быстрая предварительная фильтрация по статическим параметрам
    baseCandidates, err := r.findBaseCandidates(searchQuery)
    if err != nil {
        return nil, err
    }
    
    if len(baseCandidates) == 0 {
        return []*domain.BankDetail{}, nil
    }

    // Этап 2: Проверка динамических ограничений только для кандидатов
    finalCandidates, err := r.applyDynamicConstraints(baseCandidates, searchQuery)
    if err != nil {
        return nil, err
    }

    return finalCandidates, nil
}

// Этап 1: Быстрая фильтрация по статическим параметрам
func (r *DefaultBankDetailRepo) findBaseCandidates(searchQuery *domain.SuitablleBankDetailsQuery) ([]models.BankDetailModel, error) {
    var baseCandidates []models.BankDetailModel
    
    query := r.DB.Model(&models.BankDetailModel{}).
        Where("enabled = ?", true).
        Where("min_amount <= ? AND max_amount >= ?", searchQuery.AmountFiat, searchQuery.AmountFiat).
        Where("payment_system = ?", searchQuery.PaymentSystem).
        Where("currency = ?", searchQuery.Currency).
        Where("deleted_at IS NULL")
    
    if searchQuery.BankCode != "" {
        query = query.Where("bank_code = ?", searchQuery.BankCode)
    }
    
    if searchQuery.NspkCode != "" {
        query = query.Where("nspk_code = ?", searchQuery.NspkCode)
    }
    
    // Выбираем только ID и основные поля для быстрого отклика
    err := query.Select("id, trader_id, country, currency, inflow_currency, " +
        "min_amount, max_amount, bank_name, bank_code, nspk_code, " +
        "payment_system, delay, enabled, card_number, phone, owner, " +
        "max_orders_simultaneosly, max_amount_day, max_amount_month, " +
        "max_quantity_day, max_quantity_month, device_id, created_at, updated_at").
        Find(&baseCandidates).Error
        
    return baseCandidates, err
}

// Этап 2: Применение динамических ограничений
func (r *DefaultBankDetailRepo) applyDynamicConstraints(baseCandidates []models.BankDetailModel, searchQuery *domain.SuitablleBankDetailsQuery) ([]*domain.BankDetail, error) {
    if len(baseCandidates) == 0 {
        return []*domain.BankDetail{}, nil
    }
    
    // Получаем ID кандидатов для использования в SQL запросе
    candidateIDs := make([]string, len(baseCandidates))
    for i, candidate := range baseCandidates {
        candidateIDs[i] = candidate.ID
    }
    
    var finalCandidates []models.BankDetailModel
    
    now := time.Now()
    startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
    startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
    
    // Оптимизированный SQL запрос с использованием CTE (Common Table Expression)
    sqlQuery := `
        WITH candidate_stats AS (
            SELECT 
                bd.id,
                bd.max_orders_simultaneosly,
                bd.max_quantity_day,
                bd.max_amount_day,
                bd.max_quantity_month,
                bd.max_amount_month,
                bd.delay,
                -- Статистика по активным заявкам
                COALESCE(pending_orders.count, 0) as pending_count,
                -- Статистика за день
                COALESCE(day_stats.count, 0) as day_count,
                COALESCE(day_stats.amount, 0) as day_amount,
                -- Статистика за месяц
                COALESCE(month_stats.count, 0) as month_count,
                COALESCE(month_stats.amount, 0) as month_amount,
                -- Время последней завершенной заявки
                last_completed.last_time as last_completed_time
            FROM bank_detail_models bd
            LEFT JOIN (
                SELECT bank_details_id, COUNT(*) as count 
                FROM order_models 
                WHERE status = ? AND bank_details_id IN (?)
                GROUP BY bank_details_id
            ) pending_orders ON pending_orders.bank_details_id = bd.id
            LEFT JOIN (
                SELECT bank_details_id, COUNT(*) as count, COALESCE(SUM(amount_fiat), 0) as amount
                FROM order_models 
                WHERE status IN (?, ?) AND created_at >= ? AND bank_details_id IN (?)
                GROUP BY bank_details_id
            ) day_stats ON day_stats.bank_details_id = bd.id
            LEFT JOIN (
                SELECT bank_details_id, COUNT(*) as count, COALESCE(SUM(amount_fiat), 0) as amount
                FROM order_models 
                WHERE status IN (?, ?) AND created_at >= ? AND bank_details_id IN (?)
                GROUP BY bank_details_id
            ) month_stats ON month_stats.bank_details_id = bd.id
            LEFT JOIN (
                SELECT bank_details_id, MAX(created_at) as last_time
                FROM order_models 
                WHERE status = ? AND bank_details_id IN (?)
                GROUP BY bank_details_id
            ) last_completed ON last_completed.bank_details_id = bd.id
            WHERE bd.id IN (?)
        )
        SELECT bd.* FROM bank_detail_models bd
        JOIN candidate_stats cs ON bd.id = cs.id
        WHERE cs.pending_count < cs.max_orders_simultaneosly
          AND cs.day_count + 1 <= cs.max_quantity_day
          AND cs.day_amount + ? <= cs.max_amount_day
          AND cs.month_count + 1 <= cs.max_quantity_month
          AND cs.month_amount + ? <= cs.max_amount_month
          AND (cs.last_completed_time IS NULL OR cs.last_completed_time <= NOW() - (cs.delay / 1000000000.0) * INTERVAL '1 SECOND')
          AND NOT EXISTS (
            SELECT 1 FROM order_models om 
            WHERE om.bank_details_id = bd.id 
              AND om.status = ? 
              AND om.amount_fiat = ?
          )
    `
    
    // Выполняем оптимизированный запрос
    err := r.DB.Raw(sqlQuery,
        domain.StatusPending, candidateIDs, // pending_orders
        domain.StatusPending, domain.StatusCompleted, startOfDay, candidateIDs, // day_stats
        domain.StatusPending, domain.StatusCompleted, startOfMonth, candidateIDs, // month_stats
        domain.StatusCompleted, candidateIDs, // last_completed
        candidateIDs, // основной WHERE
        searchQuery.AmountFiat, searchQuery.AmountFiat, // суммы для проверки лимитов
        domain.StatusPending, searchQuery.AmountFiat, // NOT EXISTS
    ).Scan(&finalCandidates).Error
    
    if err != nil {
        return nil, err
    }
    
    // Преобразование в доменные объекты
    bankDetails := make([]*domain.BankDetail, len(finalCandidates))
    for i, bankDetail := range finalCandidates {
        bankDetails[i] = mappers.ToDomainBankDetail(&bankDetail)
    }
    
    return bankDetails, nil
}

func (r *DefaultBankDetailRepo) GetBankDetailsStatsByTraderID(traderID string) ([]*domain.BankDetailStat, error) {
	var bankDetails []models.BankDetailModel

	if err := r.DB.Where("trader_id = ? AND deleted_at IS NULL", traderID).Find(&bankDetails).Error; err != nil {
		return nil, err
	}

	today := time.Now().Truncate(24 * time.Hour)
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())

	stats := make([]*domain.BankDetailStat, 0, len(bankDetails))

	for _, bd := range bankDetails {
		var dayCount, monthCount int64
		var daySum, monthSum float64

		// Кол-во и сумма заявок за сегодня
		_ = r.DB.Model(&models.OrderModel{}).
			Where("bank_details_id = ? AND status IN (?) AND created_at >= ?", bd.ID, []string{string(domain.StatusCompleted), string(domain.StatusPending)}, today).
			Count(&dayCount).Error

		_ = r.DB.Model(&models.OrderModel{}).
			Select("COALESCE(SUM(amount_fiat), 0)").
			Where("bank_details_id = ? AND status IN (?) AND created_at >= ?", bd.ID, []string{string(domain.StatusCompleted), string(domain.StatusPending)}, today).
			Scan(&daySum).Error

		// Кол-во и сумма заявок за месяц
		_ = r.DB.Model(&models.OrderModel{}).
			Where("bank_details_id = ? AND status IN (?) AND created_at >= ?", bd.ID, []string{string(domain.StatusCompleted), string(domain.StatusPending)}, monthStart).
			Count(&monthCount).Error

		_ = r.DB.Model(&models.OrderModel{}).
			Select("COALESCE(SUM(amount_fiat), 0)").
			Where("bank_details_id = ? AND status IN (?) AND created_at >= ?", bd.ID, []string{string(domain.StatusCompleted), string(domain.StatusPending)}, monthStart).
			Scan(&monthSum).Error

		stats = append(stats, &domain.BankDetailStat{
			BankDetailID:      bd.ID,
			CurrentCountToday: int(dayCount),
			CurrentCountMonth: int(monthCount),
			CurrentAmountToday: daySum,
			CurrentAmountMonth: monthSum,
		})
	}

	return stats, nil
}

func (r *DefaultBankDetailRepo) GetBankDetails(filter domain.GetBankDetailsFilter) ([]*domain.BankDetail, int64, error) {
	query := r.DB.Model(&models.BankDetailModel{})

	if filter.TraderID != nil {
		query = query.Where("trader_id = ?", *filter.TraderID)
	}
	if filter.BankCode != nil {
		query = query.Where("bank_code = ?", *filter.BankCode)
	}
	if filter.Enabled != nil {
		query = query.Where("enabled = ?", *filter.Enabled)
	}
	if filter.PaymentSystem != nil {
		query = query.Where("payment_system = ?", *filter.PaymentSystem)
	}
	if filter.BankDetailID != nil {
		query = query.Where("id = ?", *filter.BankDetailID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count failed: %w", err)
	}

	offset := (filter.Page - 1) * filter.Limit
	query = query.Offset(offset).Limit(filter.Limit)

	var bankDetailModels []models.BankDetailModel
	if err := query.Find(&bankDetailModels).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find bank detail models: %w", err)
	}

	bankDetails := make([]*domain.BankDetail, len(bankDetailModels))
	for i, bankDetailModel := range bankDetailModels {
		bankDetails[i] = mappers.ToDomainBankDetail(&bankDetailModel)
	}

	return bankDetails, total, nil
}